package firestore

import (
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/keys-pub/keys/ds"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

type docsIterator struct {
	iter     *firestore.DocumentIterator
	prefix   string
	parent   string
	pathOnly bool
}

func (i *docsIterator) Next() (*ds.Document, error) {
	if i.iter == nil {
		return nil, nil
	}
	doc, err := i.iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	k := doc.Ref.ID
	if i.prefix != "" && !strings.HasPrefix(k, i.prefix) {
		// We've reached an entry not matching prefix, so end iteration
		// TODO: Is there a more efficient way to do this in the query?
		return nil, nil
	}
	kp := ds.Path(i.parent, k)

	if i.pathOnly {
		out := ds.NewDocument(kp, nil)
		out.CreatedAt = doc.CreateTime
		out.UpdatedAt = doc.UpdateTime
		return out, nil
	}

	m := doc.Data()
	b, ok := m["data"].([]byte)
	if !ok {
		return nil, errors.Errorf("firestore value missing data")
	}
	out := ds.NewDocument(kp, b)
	out.CreatedAt = doc.CreateTime
	out.UpdatedAt = doc.UpdateTime
	return out, nil
}

func (i *docsIterator) Release() {
	if i.iter != nil {
		i.iter.Stop()
	}
}

type colsIterator struct {
	iter *firestore.CollectionIterator
}

func (i *colsIterator) Next() (*ds.Collection, error) {
	col, err := i.iter.Next()
	if err == iterator.Done {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ds.Collection{Path: ds.Path(col.ID)}, nil
}

func (i *colsIterator) Release() {
	// Nothing to do for firestore.CollectionIterator
}
