package screen

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

const (
	cursorIcon = rune('|')
)

type Screen struct {
	mu        sync.Mutex
	buffer    []rune
	screen    tcell.Screen
	cursorIdx int
}

func NewScreen() (Screen, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return Screen{}, err
	}

	return Screen{
		mu:        sync.Mutex{},
		buffer:    []rune{},
		cursorIdx: 0,
		screen:    s,
	}, nil
}

func (s *Screen) Init() error {
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

func (s *Screen) Content() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	content := string(s.buffer)
	return content
}

func (s *Screen) DeleteChunk(idx, length int) {
	s.mu.Lock()
	if idx < 0 || idx+length > len(s.buffer) {
		panic("deleting not existing chunk")
	}

	s.buffer = append(s.buffer[:idx], s.buffer[idx+length:]...)
	s.moveCursor(-length)
	s.mu.Unlock()

	s.render()
}

func (s *Screen) InsertChunk(idx int, chunk string) {
	s.mu.Lock()
	if idx < 0 {
		idx = 0
	}

	newBuffer := make([]rune, 0, len(s.buffer)+len(chunk))
	newBuffer = append(newBuffer, s.buffer[:idx]...)
	newBuffer = append(newBuffer, []rune(chunk)...)
	newBuffer = append(newBuffer, s.buffer[idx:]...)
	s.buffer = newBuffer

	if idx <= s.cursorIdx {
		s.moveCursor(len(chunk))
	}
	s.mu.Unlock()

	s.render()
}

func (s *Screen) backspace() {
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

func (s *Screen) moveCursor(pos int) {
	s.cursorIdx += pos
	if s.cursorIdx > len(s.buffer) {
		s.cursorIdx = len(s.buffer)
	}
	if s.cursorIdx < 0 {
		s.cursorIdx = 0
	}
}

func (s *Screen) key(r rune) {
	s.buffer = append(s.buffer, r)
	last := len(s.buffer) - 1
	for i := last; i > 0 && i > s.cursorIdx; i-- {
		s.buffer[i], s.buffer[i-1] = s.buffer[i-1], s.buffer[i]
	}

	s.moveCursor(1)
}

func (s *Screen) contentWithCursor() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	content := make([]rune, 0, len(s.buffer)+1)

	for i, r := range s.buffer {
		if i == s.cursorIdx {
			content = append(content, cursorIcon)
		}
		content = append(content, r)
	}

	if s.cursorIdx == len(s.buffer) {
		content = append(content, cursorIcon)
	}

	return string(content)
}

func (s *Screen) render() {
	content := s.contentWithCursor()
	s.mu.Lock()
	defer s.mu.Unlock()

	s.screen.Clear()
	for i, r := range content {
		s.screen.SetContent(i, 0, r, nil, tcell.StyleDefault)
	}
	s.screen.Show()
}
