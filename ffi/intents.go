package ffi

import "go.rockorager.dev/vaxis/ui"

// stringIntent is a string-typed intent. Ard cannot define a Go interface
// method on a string newtype, so the ui.Intent implementation lives here; the
// intent names, key bindings, and handlers are all pure Ard.
type stringIntent string

func (i stringIntent) IntentType() ui.IntentType { return ui.IntentType(i) }

// Intent wraps an intent type as a ui.Intent value for Shortcuts key maps.
func Intent(t ui.IntentType) ui.Intent { return stringIntent(t) }
