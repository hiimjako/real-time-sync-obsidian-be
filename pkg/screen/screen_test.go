package screen

import (
	"testing"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestScreen(t *testing.T) {
	s, err := NewScreen()
	assert.NoError(t, err)

	simulationScreen := tcell.NewSimulationScreen("UTF-8")
	s.screen = simulationScreen

	//nolint:errcheck
	go s.Init()

	time.Sleep(10 * time.Millisecond)

	sendKey(t, s.screen, tcell.KeyRune, 'h')
	sendKey(t, s.screen, tcell.KeyRune, 'x')
	sendKey(t, s.screen, tcell.KeyRune, 'l')
	sendKey(t, s.screen, tcell.KeyRune, 'o')

	sendKey(t, s.screen, tcell.KeyLeft, ' ')
	sendKey(t, s.screen, tcell.KeyLeft, ' ')
	sendKey(t, s.screen, tcell.KeyBackspace, ' ')
	sendKey(t, s.screen, tcell.KeyRune, 'e')
	sendKey(t, s.screen, tcell.KeyRune, 'l')

	sendKey(t, s.screen, tcell.KeyRight, ' ')
	sendKey(t, s.screen, tcell.KeyRight, ' ')
	sendKey(t, s.screen, tcell.KeyRune, '!')

	assert.Contains(t, getDisplayedContent(simulationScreen), "hello!|")
	assert.Equal(t, s.Content(), "hello!")

	s.DeleteChunk(2, 2)

	assert.Contains(t, getDisplayedContent(simulationScreen), "heo!|")
	assert.Equal(t, s.Content(), "heo!")
}

func sendKey(t testing.TB, s tcell.Screen, key tcell.Key, r rune) {
	assert.NoError(t, s.PostEvent(tcell.NewEventKey(key, r, tcell.ModNone)))
	time.Sleep(1 * time.Millisecond)
}

func getDisplayedContent(s tcell.SimulationScreen) string {
	content := ""

	contents, _, _ := s.GetContents()
	for _, c := range contents {
		aRune, _ := utf8.DecodeRune(c.Bytes)
		content += string(aRune)
	}

	return content
}
