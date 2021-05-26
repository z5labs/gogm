// Copyright (c) 2021 MindStand Technologies, Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package gogm

import (
	_neoLog "github.com/neo4j/neo4j-go-driver/v4/neo4j/log"
	"log"
)

type Logger interface {
	Debug(s string)
	Debugf(s string, vals ...interface{})

	Info(s string)
	Infof(s string, vals ...interface{})

	Warn(s string)
	Warnf(s string, vals ...interface{})

	Error(s string)
	Errorf(s string, vals ...interface{})

	Fatal(s string)
	Fatalf(s string, vals ...interface{})
}

type wrapNeoLogger struct {
	log Logger
}

func wrapLogger(_log Logger) _neoLog.Logger {
	return &wrapNeoLogger{
		log: _log,
	}
}

func (wn *wrapNeoLogger) Error(name string, id string, err error) {
	wn.log.Errorf("[name=%s] [id=%s] [err=%v]", name, id, err)
}
func (wn *wrapNeoLogger) Warnf(name string, id string, msg string, args ...interface{}) {
	arr := []interface{}{name, id}
	arr = append(arr, args...)
	wn.log.Warnf("[name=%s] [id=%s] "+msg, arr...)
}
func (wn *wrapNeoLogger) Infof(name string, id string, msg string, args ...interface{}) {
	arr := []interface{}{name, id}
	arr = append(arr, args...)
	wn.log.Infof("[name=%s] [id=%s] "+msg, arr...)
}
func (wn *wrapNeoLogger) Debugf(name string, id string, msg string, args ...interface{}) {
	arr := []interface{}{name, id}
	arr = append(arr, args...)
	wn.log.Debugf("[name=%s] [id=%s] "+msg, arr...)
}

type defaultLogger struct {
}

func (d defaultLogger) Debug(s string) {
	log.Println("[DEBUG] " + s)
}

func (d defaultLogger) Debugf(s string, vals ...interface{}) {
	log.Printf("[DEBUG] "+s+"\n", vals...)
}

func (d defaultLogger) Info(s string) {
	log.Println("[INFO] " + s)
}

func (d defaultLogger) Infof(s string, vals ...interface{}) {
	log.Printf("[INFO] "+s+"\n", vals...)
}

func (d defaultLogger) Warn(s string) {
	log.Println("[WARN] " + s)
}

func (d defaultLogger) Warnf(s string, vals ...interface{}) {
	log.Printf("[WARN] "+s+"\n", vals...)
}

func (d defaultLogger) Error(s string) {
	log.Println("[ERROR] " + s)
}

func (d defaultLogger) Errorf(s string, vals ...interface{}) {
	log.Printf("[ERROR] "+s+"\n", vals...)
}

func (d defaultLogger) Fatal(s string) {
	log.Fatalln("[FATAL] " + s)
}

func (d defaultLogger) Fatalf(s string, vals ...interface{}) {
	log.Fatalf("[FATAL] "+s+"\n", vals...)
}

func GetDefaultLogger() Logger {
	return &defaultLogger{}
}
