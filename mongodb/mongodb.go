package mongodb

import (
	"sync"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/skiplee85/common/log"
)

// session
type Session struct {
	*mgo.Session
	ref   int
	index int
}

// session heap
type SessionHeap []*Session

func (h SessionHeap) Len() int {
	return len(h)
}

func (h SessionHeap) Less(i, j int) bool {
	return h[i].ref < h[j].ref
}

func (h SessionHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *SessionHeap) Push(s interface{}) {
	s.(*Session).index = len(*h)
	*h = append(*h, s.(*Session))
}

func (h *SessionHeap) Pop() interface{} {
	l := len(*h)
	s := (*h)[l-1]
	s.index = -1
	*h = (*h)[:l-1]
	return s
}

type DialContext struct {
	sync.Mutex
	// sessions SessionHeap
	sessions chan *Session
}

// goroutine safe
func Dial(url string, sessionNum int) (*DialContext, error) {
	c, err := DialWithTimeout(url, sessionNum, 10*time.Second, 5*time.Minute)
	return c, err
}

// goroutine safe
func DialWithTimeout(url string, sessionNum int, dialTimeout time.Duration, timeout time.Duration) (*DialContext, error) {
	if sessionNum <= 0 {
		sessionNum = 100
		log.Info("invalid sessionNum, reset to %v", sessionNum)
	}

	s, err := mgo.DialWithTimeout(url, dialTimeout)
	if err != nil {
		return nil, err
	}
	s.SetSyncTimeout(timeout)
	s.SetSocketTimeout(timeout)

	c := new(DialContext)

	// sessions
	c.sessions = make(chan *Session, sessionNum)
	c.sessions <- &Session{s, 0, 0}
	for i := 1; i < sessionNum; i++ {
		c.sessions <- &Session{s.New(), 0, i}
	}
	// heap.Init(&c.sessions)
	go c.poolPing()

	return c, nil
}

func (c *DialContext) poolPing() {
	defer time.AfterFunc(time.Minute, c.poolPing)

	counter := len(c.sessions)
	for i := 0; i < counter; i++ {
		s := <-c.sessions
		if err := s.Ping(); err != nil {
			s.Refresh()
			log.Error("ping error. %s", err.Error())
		}
		c.sessions <- s
	}
}

// goroutine safe
func (c *DialContext) Close() {
	c.Lock()
	for len(c.sessions) > 0 {
		s := <-c.sessions
		s.Close()
		if s.ref != 0 {
			log.Error("session ref = %v", s.ref)
		}
	}
	c.Unlock()
}

// goroutine safe
func (c *DialContext) Ref() *Session {
	// c.Lock()
	s := <-c.sessions
	// if s.ref == 0 {
	// 	if s.Ping() != nil {
	// 		s.Refresh()
	// 	}
	// }
	s.ref++
	// heap.Fix(&c.sessions, 0)
	// c.Unlock()
	return s
}

// goroutine safe
func (c *DialContext) UnRef(s *Session) {
	// c.Lock()
	s.ref--
	c.sessions <- s
	// heap.Fix(&c.sessions, s.index)
	// c.Unlock()
}

// goroutine safe
func (c *DialContext) EnsureCounter(db string, collection string, id string) error {
	s := c.Ref()
	defer c.UnRef(s)

	err := s.DB(db).C(collection).Insert(bson.M{
		"_id": id,
		"seq": 0,
	})
	if mgo.IsDup(err) {
		return nil
	} else {
		return err
	}
}

// goroutine safe
func (c *DialContext) NextSeq(db string, collection string, id string) (int, error) {
	s := c.Ref()
	defer c.UnRef(s)

	var res struct {
		Seq int
	}
	_, err := s.DB(db).C(collection).FindId(id).Apply(mgo.Change{
		Update:    bson.M{"$inc": bson.M{"seq": 1}},
		ReturnNew: true,
	}, &res)

	return res.Seq, err
}

// goroutine safe
func (c *DialContext) EnsureIndex(db string, collection string, key []string) error {
	s := c.Ref()
	defer c.UnRef(s)

	return s.DB(db).C(collection).EnsureIndex(mgo.Index{
		Key:    key,
		Unique: false,
		Sparse: true,
	})
}

// goroutine safe
func (c *DialContext) EnsureUniqueIndex(db string, collection string, key []string) error {
	s := c.Ref()
	defer c.UnRef(s)

	return s.DB(db).C(collection).EnsureIndex(mgo.Index{
		Key:    key,
		Unique: true,
		Sparse: true,
	})
}
