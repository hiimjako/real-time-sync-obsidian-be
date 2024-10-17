package screen

import (
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
)

const (
	cursorIcon = rune('|')
)

type screen struct {
	mu        sync.Mutex
	buffer    []rune
	screen    tcell.Screen
	cursorIdx int
}

func NewScreen() (screen, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return screen{}, err
	}

	return screen{
		mu:        sync.Mutex{},
		buffer:    []rune{},
		cursorIdx: 0,
		screen:    s,
	}, nil
}

func (s *screen) Init() error {
	if err := s.screen.Init(); err != nil {
		return err
	}
	defer s.screen.Fini()

	for {
		s.render()

		event := s.screen.PollEvent()
		s.mu.Lock()
		switch ev := event.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return nil
			}

			if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
				s.backspace()
			} else if ev.Key() == tcell.KeyLeft {
				s.moveCursor(-1)
			} else if ev.Key() == tcell.KeyRight {
				s.moveCursor(1)
			} else if ev.Key() == tcell.KeyRune {
				s.key(ev.Rune())
			}
		case *tcell.EventResize:
			s.screen.Sync()
		}
		s.mu.Unlock()
	}
}

func (s *screen) Content() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var content strings.Builder
	content.Grow(len(s.buffer))

	for i := 0; i < len(s.buffer); i++ {
		r, _, _, _ := s.screen.GetContent(i, 0)
		content.WriteRune(r)
	}

	return content.String()
}

func (s *screen) backspace() {
	if s.cursorIdx == 0 {
		return
	}

	if s.cursorIdx == len(s.buffer) {
		s.buffer = s.buffer[:s.cursorIdx-1]
	} else {
		s.buffer = append(s.buffer[:s.cursorIdx-1], s.buffer[s.cursorIdx:]...)
	}
	s.moveCursor(-1)
}

func (s *screen) moveCursor(pos int) {
	s.cursorIdx += pos
	if s.cursorIdx > len(s.buffer) {
		s.cursorIdx = len(s.buffer)
	}
	if s.cursorIdx < 0 {
		s.cursorIdx = 0
	}
}

func (s *screen) key(r rune) {
	s.buffer = append(s.buffer, r)
	last := len(s.buffer) - 1
	for i := last; i > 0 && i > s.cursorIdx; i-- {
		s.buffer[i], s.buffer[i-1] = s.buffer[i-1], s.buffer[i]
	}

	s.moveCursor(1)
}

func (s *screen) render() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.screen.Clear()
	cursorOffset := 0
	for i, r := range s.buffer {
		if i == s.cursorIdx {
			s.screen.SetContent(i, 0, cursorIcon, nil, tcell.StyleDefault)
			cursorOffset = 1
		}
		s.screen.SetContent(i+cursorOffset, 0, r, nil, tcell.StyleDefault)
	}
	if cursorOffset == 0 {
		s.screen.SetContent(len(s.buffer), 0, cursorIcon, nil, tcell.StyleDefault)
	}
	s.screen.Show()
}
