package common

// CodeReview to be updated.
type CodeReview struct {
	ID  string
	Ref string
}

// Empty checks if CodeReview is empty.
func (v CodeReview) Empty() bool {
	return v.ID == "" || v.Ref == ""
}
