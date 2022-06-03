// Copyright (c) 2022 MindStand Technologies, Inc
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
	"context"
	"fmt"
	"github.com/cornelk/hashmap"
)

//drops all known indexes
func dropAllIndexesAndConstraints(ctx context.Context, gogm *Gogm) error {
	if gogm.boltMajorVersion >= 4 {
		for _, db := range gogm.config.TargetDbs {
			err := dropAllIndexesAndConstraintsV4(ctx, gogm, db)
			if err != nil {
				return fmt.Errorf("failed to drop indexes and constraints for db %s on db version 4+, %w", db, err)
			}
		}
	} else {
		err := dropAllIndexesAndConstraintsV3(ctx, gogm)
		if err != nil {
			return fmt.Errorf("failed to drop indexes and constraints on db version 3, %w", err)
		}
	}

	return nil
}

//creates all indexes
func createAllIndexesAndConstraints(ctx context.Context, gogm *Gogm, mappedTypes *hashmap.HashMap) error {
	if gogm.boltMajorVersion >= 4 {
		for _, db := range gogm.config.TargetDbs {
			err := createAllIndexesAndConstraintsV4(ctx, gogm, mappedTypes, db)
			if err != nil {
				return fmt.Errorf("failed to create indexes and constraints for db %s on db version 4+, %w", db, err)
			}
		}
	} else {
		err := createAllIndexesAndConstraintsV3(ctx, gogm, mappedTypes)
		if err != nil {
			return fmt.Errorf("failed to create indexes and constraints on db version 3, %w", err)
		}
	}
	return nil
}

//verifies all indexes
func verifyAllIndexesAndConstraints(ctx context.Context, gogm *Gogm, mappedTypes *hashmap.HashMap) error {
	if gogm.boltMajorVersion >= 4 {
		for _, db := range gogm.config.TargetDbs {
			err := verifyAllIndexesAndConstraintsV4(ctx, gogm, mappedTypes, db)
			if err != nil {
				return fmt.Errorf("failed to verify indexes and constraints for db %s on db version 4+, %w", db, err)
			}
		}
	} else {
		err := verifyAllIndexesAndConstraintsV3(ctx, gogm, mappedTypes)
		if err != nil {
			return fmt.Errorf("failed to verify all indexes and contraints on db version 3, %w", err)
		}
	}

	return nil
}
