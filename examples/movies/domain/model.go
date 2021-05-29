package domain

import "github.com/mindstand/gogm/v2"

type Movie struct {
	gogm.BaseNode

	Title        string     `gogm:"name=title"`
	ReleasedYear int        `gogm:"name=released"`
	TagLine      string     `gogm:"name=tagline"`
	Actors       []*ActedIn `gogm:"direction=incoming;relationship=ACTED_IN"`
	Directors    []*Person  `gogm:"direction=incoming;relationship=DIRECTED"`
	Producers    []*Person  `gogm:"direction=incoming;relationship=PRODUCED"`
	Followers    []*Person  `gogm:"direction=incoming;relationship=FOLLOWS"`
	Writers      []*Person  `gogm:"direction=incoming;relationship=WROTE"`
	Reviewers    []*Person  `gogm:"direction=incoming;relationship=REVIEWED"`
}

type Person struct {
	gogm.BaseNode

	Name     string     `gogm:"name=name"`
	BornYear int        `gogm:"name=born"`
	Directed []*Movie   `gogm:"direction=outgoing;relationship=DIRECTED"`
	Produced []*Movie   `gogm:"direction=outgoing;relationship=PRODUCED"`
	Follows  []*Person  `gogm:"direction=outgoing;relationship=FOLLOWS"`
	Wrote    []*Movie   `gogm:"direction=outgoing;relationship=WROTE"`
	Reviewed []*Movie   `gogm:"direction=outgoing;relationship=REVIEWED"`
	ActedIn  []*ActedIn `gogm:"direction=outgoing;relationship=ACTED_IN"`
}

type ActedIn struct {
	gogm.BaseNode

	Person *Person
	Movie  *Movie
	Roles  []string `gogm:"name=roles;properties"`
}

