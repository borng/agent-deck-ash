package ui

import "strconv"

// Empty metadata means no explicit slot, not a claim about the running login.
func storedAccountLabel(account string) string {
	if account == "" {
		return "inherited"
	}
	return strconv.Quote(account)
}

const storedAccountPrefix = " [account:"

// Immutable presentation travels with the raw slot in the render snapshot.
// Width-independent quoting is shared across rows with the same stored slot.
type accountPresentation struct {
	label  string
	badge  string
	width  int
	quoted bool
}

func newAccountPresentation(account string) accountPresentation {
	label := storedAccountLabel(account)
	// No stored slot is the common case and says nothing about the running
	// login, so the row badge would be pure noise on every session. The label
	// still travels in the snapshot for the detail card, which has the room to
	// explain "inherited" and is the place a user goes to ask the question.
	if account == "" {
		return accountPresentation{label: label}
	}
	return accountPresentation{
		label:  label,
		badge:  storedAccountPrefix + label + "]",
		width:  len(storedAccountPrefix) + cellWidth(label) + 1,
		quoted: true,
	}
}

// Quote before styling/truncation so terminal controls cannot become commands.
// Keep the delimiters visible even when a long account needs an ellipsis.
func (p accountPresentation) fit(budget int) (string, int) {
	if p.width <= budget {
		return p.badge, p.width
	}
	available := budget - len(storedAccountPrefix) - 1
	if available < 1 {
		return "", 0
	}
	label := p.label
	if p.quoted {
		if available < 3 {
			return "", 0
		}
		label = "\"" + cellTruncate(label[1:len(label)-1], available-2, "…") + "\""
	} else {
		label = cellTruncate(label, available, "…")
	}
	return storedAccountPrefix + label + "]", len(storedAccountPrefix) + cellWidth(label) + 1
}
