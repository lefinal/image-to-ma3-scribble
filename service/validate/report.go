// Package validate implements a validation framework. Reporter is used as
// syntactic sugar in validation. Set the next field using NextField and then
// report errors with Report. The final error list can be retrieved via
// ErrorList.
package validate

import (
	"encoding/json"
	"fmt"
	"github.com/lefinal/meh"
	"strings"
)

// Issue represents a validation issue.
//
//nolint:errname
type Issue struct {
	Field    string
	BadValue any
	Detail   string
}

func (issue Issue) String() string {
	switch v := issue.BadValue.(type) {
	case []byte:
		issue.BadValue = string(v)
	case json.RawMessage:
		issue.BadValue = string(v)
	}
	return fmt.Sprintf("%s: %v - %s", issue.Field, issue.BadValue, issue.Detail)
}

// Error returns the string-representation for the Issue via String.
func (issue Issue) Error() string {
	return issue.String()
}

// IssueList is an Issue-collection that implements the error-interface. This can
// be returned and used in the HTTP API for responding with structured errors.
//
//nolint:errname
type IssueList struct {
	Issues []Issue
}

// Error returns the error-string for the first issue.
func (issue IssueList) Error() string {
	firstIssue := Issue{Detail: "unknown"} //nolint:exhaustruct
	if len(issue.Issues) > 0 {
		firstIssue = issue.Issues[0]
	}
	return firstIssue.Error()
}

// NewIssue creates a new Issue with the specified field, bad value, and detail.
func NewIssue(field *Path, badValue any, detail string) Issue {
	return Issue{
		// Trim empty start segment from the field path.
		Field:    strings.TrimPrefix(field.String(), "[]."),
		BadValue: badValue,
		Detail:   detail,
	}
}

// Report represents the validation report.
type Report struct {
	Warnings []Issue
	Errors   []Issue
}

// NewReport creates a new empty Report.
func NewReport() *Report {
	return &Report{
		Warnings: make([]Issue, 0),
		Errors:   make([]Issue, 0),
	}
}

// AddWarning the given warning for the last field that was set via NextField.
func (r *Report) AddWarning(issue Issue) {
	r.Warnings = append(r.Warnings, issue)
}

// AddError the given error for the last field that was set via NextField.
func (r *Report) AddError(issue Issue) {
	r.Errors = append(r.Errors, issue)
}

// AddReport appends the warnings and errors from another Report to the current
// Report.
func (r *Report) AddReport(otherReport *Report) {
	r.Warnings = append(r.Warnings, otherReport.Warnings...)
	r.Errors = append(r.Errors, otherReport.Errors...)
}

// Err returns an error if there are any errors in the report. It returns an
// IssueList, wrapped with a meh.ErrBadInput-code and the errors and warnings as
// details.
func (r *Report) Err() error {
	if len(r.Errors) == 0 {
		return nil
	}
	err := error(IssueList{Issues: r.Errors})
	err = meh.ApplyCode(err, meh.ErrBadInput)
	err = meh.ApplyDetails(err, meh.Details{
		"errors":   r.Errors,
		"warnings": r.Warnings,
	})
	return err
}

// Reporter is used as syntactic sugar in validation. Set the next field using
// NextField and then report errors with Report. The final error list can be
// retrieved via ErrorList.
type Reporter struct {
	fieldPath  *Path
	fieldValue any
	report     *Report
}

// NextField sets the field that calls to Error and Warn will use.
func (r *Reporter) NextField(fieldPath *Path, fieldValue any) {
	r.fieldPath = fieldPath
	r.fieldValue = fieldValue
}

// CurrentFieldPath returns the current Path.
func (r *Reporter) CurrentFieldPath() *Path {
	return r.fieldPath
}

// Warn the given warning for the last field that was set via NextField.
func (r *Reporter) Warn(warnMsg string) {
	r.report.Warnings = append(r.report.Warnings, NewIssue(r.fieldPath, r.fieldValue, warnMsg))
}

// Error the given error for the last field that was set via NextField.
func (r *Reporter) Error(errMsg string) {
	r.report.Errors = append(r.report.Errors, NewIssue(r.fieldPath, r.fieldValue, errMsg))
}

// AddReport adds the given Report.
func (r *Reporter) AddReport(otherReport *Report) {
	r.report.AddReport(otherReport)
}

// Report returns the final Report that contains all issues.
func (r *Reporter) Report() *Report {
	return r.report
}

// NewReporter creates a new Reporter that is ready to use.
func NewReporter() *Reporter {
	return &Reporter{
		fieldPath:  nil,
		fieldValue: nil,
		report:     NewReport(),
	}
}
