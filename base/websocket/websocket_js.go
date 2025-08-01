// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build js

package websocket

import (
	"syscall/js"

	"github.com/cogentcore/webgpu/jsx"
)

// Client represents a WebSocket client connection.
// You can use [Connect] to create a new Client.
type Client struct {

	// ws is the underlying JavaScript WebSocket object.
	// See https://developer.mozilla.org/en-US/docs/Web/API/WebSocket
	ws js.Value
}

// Connect connects to a WebSocket server and returns a [Client].
func Connect(url string) (*Client, error) {
	ws := js.Global().Get("WebSocket").New(url)
	ws.Set("binaryType", "arraybuffer")
	return &Client{ws: ws}, nil
}

// OnMessage sets a callback function to be called when a message is received.
// This function can only be called once on native.
func (c *Client) OnMessage(f func(typ MessageTypes, msg []byte)) {
	c.ws.Call("addEventListener", "message", js.FuncOf(func(this js.Value, args []js.Value) any {
		data := args[0].Get("data")
		if data.Type() == js.TypeString {
			f(TextMessage, []byte(data.String()))
			return nil
		}
		array := js.Global().Get("Uint8ClampedArray").New(data)
		b := make([]byte, array.Length())
		js.CopyBytesToGo(b, array) // TODO: more performant way to do this, perhaps with gopherjs/goscript?
		f(BinaryMessage, b)
		return nil
	}))
}

// Send sends a message to the WebSocket server with the given type and message.
func (c *Client) Send(typ MessageTypes, msg []byte) error {
	if typ == TextMessage {
		c.ws.Call("send", string(msg))
		return nil
	}
	array := jsx.BytesToJS(msg)
	c.ws.Call("send", array)
	return nil
}

// Close cleanly closes the WebSocket connection.
// It does not directly trigger [Client.OnClose], but once the connection
// is closed, that will trigger it.
func (c *Client) Close() error {
	c.ws.Call("close")
	return nil
}

// OnClose sets a callback function to be called when the connection is closed.
// This function can only be called once on native.
func (c *Client) OnClose(f func()) {
	c.ws.Call("addEventListener", "close", js.FuncOf(func(this js.Value, args []js.Value) any {
		f()
		return nil
	}))
}
