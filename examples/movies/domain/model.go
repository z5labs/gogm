package domain

import (
	"fmt"
	"github.com/mindstand/gogm/v2"
	"reflect"
)

type Movie struct {
	gogm.BaseNode

	Title        string `gogm:"name=title" json:"title"`
	ReleasedYear int    `gogm:"name=released" json:"released_year"`
	TagLine      string `gogm:"name=tagline" json:"tag_line"`

	Actors    []*ActedInEdge `gogm:"direction=incoming;relationship=ACTED_IN" json:"actors"`
	Directors []*Person      `gogm:"direction=incoming;relationship=DIRECTED" json:"directors"`
	Producers []*Person      `gogm:"direction=incoming;relationship=PRODUCED" json:"producers"`
	Followers []*Person      `gogm:"direction=incoming;relationship=FOLLOWS" json:"followers"`
	Writers   []*Person      `gogm:"direction=incoming;relationship=WROTE" json:"writers"`
	Reviewers []*Person      `gogm:"direction=incoming;relationship=REVIEWED" json:"reviewers"`
}

type Person struct {
	gogm.BaseNode

	Name     string         `gogm:"name=name" json:"name"`
	BornYear int            `gogm:"name=born" json:"born_year"`
	Directed []*Movie       `gogm:"direction=outgoing;relationship=DIRECTED" json:"-"`
	Produced []*Movie       `gogm:"direction=outgoing;relationship=PRODUCED" json:"-"`
	Follows  []*Movie       `gogm:"direction=outgoing;relationship=FOLLOWS" json:"-"`
	Wrote    []*Movie       `gogm:"direction=outgoing;relationship=WROTE" json:"-"`
	Reviewed []*Movie       `gogm:"direction=outgoing;relationship=REVIEWED" json:"-"`
	ActedIn  []*ActedInEdge `gogm:"direction=outgoing;relationship=ACTED_IN" json:"-"`
}

type ActedInEdge struct {
	gogm.BaseNode

	Start *Person  `json:"person"`
	End   *Movie   `json:"-"`
	Roles []string `gogm:"name=roles;properties"`
}

func (a *ActedInEdge) GetStartNode() interface{} {
	return a.Start
}

func (a *ActedInEdge) GetStartNodeType() reflect.Type {
	return reflect.TypeOf(&Person{})
}

func (a *ActedInEdge) SetStartNode(v interface{}) error {
	s, ok := v.(*Person)
	if !ok {
		return fmt.Errorf("cannot cast %T to *Person", s)
	}

	a.Start = s
	return nil
}

func (a *ActedInEdge) GetEndNode() interface{} {
	return a.End
}

func (a *ActedInEdge) GetEndNodeType() reflect.Type {
	return reflect.TypeOf(&Movie{})
}

func (a *ActedInEdge) SetEndNode(v interface{}) error {
	e, ok := v.(*Movie)
	if !ok {
		return fmt.Errorf("cannot cast %T to *Movie", e)
	}

	a.End = e
	return nil
}
