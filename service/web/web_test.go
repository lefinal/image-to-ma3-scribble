package web

import (
	"github.com/lefinal/image-to-ma3-scribble/validate"
	"github.com/lefinal/meh"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_invalidFieldsForError(t *testing.T) {
	sampleIssueList := func() validate.IssueList {
		return validate.IssueList{
			Issues: []validate.Issue{
				{
					Field:    "my-field",
					BadValue: "my-value",
					Detail:   "my-detail",
				},
			},
		}
	}

	tests := []struct {
		name   string
		e      func() error
		expect []errorResponseInvalidField
	}{
		{
			name: "nil",
			e: func() error {
				return nil
			},
			expect: nil,
		},
		{
			name: "neutral",
			e: func() error {
				return &meh.Error{Code: meh.ErrNeutral}
			},
			expect: nil,
		},
		{
			name: "internal",
			e: func() error {
				return meh.NewInternalErr("sad life", nil)
			},
			expect: nil,
		},
		{
			name: "internal with issue list",
			e: func() error {
				issueList := sampleIssueList()
				return meh.NewInternalErrFromErr(issueList, "sad life", nil)
			},
			expect: nil,
		},
		{
			name: "bad input with issue list",
			e: func() error {
				issueList := sampleIssueList()
				return meh.NewBadInputErrFromErr(issueList, "sad life", nil)
			},
			expect: []errorResponseInvalidField{
				{
					Field:          "my-field",
					Message:        "my-detail",
					Code:           "invalid",
					ValidationCode: "",
					Arguments:      nil,
				},
			},
		},
		{
			name: "internal with bad input with issue list",
			e: func() error {
				issueList := sampleIssueList()
				e := meh.NewBadInputErrFromErr(issueList, "sad bad input life", nil)
				return meh.NewInternalErrFromErr(e, "sad internal life", nil)
			},
			expect: nil,
		},
		{
			name: "bad input with internal with issue list",
			e: func() error {
				issueList := sampleIssueList()
				e := meh.NewInternalErrFromErr(issueList, "sad internal life", nil)
				return meh.NewBadInputErrFromErr(e, "sad bad input life", nil)
			},
			expect: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := invalidFieldsForError(tt.e())
			if tt.expect == nil {
				assert.Nil(t, got)
			} else {
				assert.ElementsMatch(t, tt.expect, got)
			}
		})
	}
}
