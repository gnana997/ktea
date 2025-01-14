package subjects_page

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/sradmin"
	"ktea/tests"
	"ktea/tests/keys"
	"ktea/ui"
	"testing"
)

type MockSubjectsLister struct {
	subjectListingStartedMsg sradmin.SubjectListingStartedMsg
}

func (m *MockSubjectsLister) ListSubjects() tea.Msg {
	return nil
}

type MockSubjectsDeleter struct {
	deletionResultMsg tea.Msg
}

type DeletedSubjectMsg struct {
	Subject string
	Version int
}

func (m *MockSubjectsDeleter) DeleteSubject(subject string) tea.Msg {
	return m.deletionResultMsg
}

func TestSubjectsPage(t *testing.T) {

	t.Run("Render listed subjects and number of versions", func(t *testing.T) {

		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockSubjectsDeleter{},
		)

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 10; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:     fmt.Sprintf("subject%d", i),
					Versions: versions,
				})
		}
		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		render := subjectsPage.View(ui.TestKontext, ui.TestRenderer)

		for i := 0; i < 10; i++ {
			assert.Regexp(t, fmt.Sprintf("subject%d\\W+%d", i, i+1), render)
		}
	})

	t.Run("Remove delete subject from table", func(t *testing.T) {

		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockSubjectsDeleter{},
		)

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 10; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:     fmt.Sprintf("subject%d", i),
					Versions: versions,
				})
		}
		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		subjectsPage.Update(sradmin.SubjectDeletedMsg{SubjectName: subjects[4].Name})

		render := subjectsPage.View(ui.TestKontext, ui.TestRenderer)

		assert.NotRegexp(t, "subject4\\W+5", render)
	})

	t.Run("When deletion spinner active do not allow other cmdbars to activate", func(t *testing.T) {

		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockSubjectsDeleter{},
		)
		subjectsPage.Update(sradmin.SubjectDeletionStartedMsg{})

		subjectsPage.Update(keys.Key('/'))

		render := subjectsPage.View(ui.TestKontext, ui.TestRenderer)

		assert.Contains(t, render, " ⏳ Deleting Subject")
	})

	t.Run("Delete subject", func(t *testing.T) {

		deleter := MockSubjectsDeleter{}
		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&deleter,
		)

		var subjects []sradmin.Subject
		var versions []int
		for i := 0; i < 10; i++ {
			versions = append(versions, i)
			subjects = append(subjects,
				sradmin.Subject{
					Name:     fmt.Sprintf("subject%d", i),
					Versions: versions,
				})
		}
		subjectsPage.Update(sradmin.SubjectsListedMsg{Subjects: subjects})

		// render so the table's first row is selected
		render := subjectsPage.View(ui.NewTestKontext(), ui.TestRenderer)
		assert.NotRegexp(t, "┃ 🗑️  subject1 will be deleted permanently\\W+Delete!\\W+Cancel.", render)

		t.Run("F2 triggers subject delete", func(t *testing.T) {
			subjectsPage.Update(keys.Key(tea.KeyDown))
			subjectsPage.Update(keys.Key(tea.KeyF2))

			render = subjectsPage.View(ui.NewTestKontext(), ui.TestRenderer)

			assert.Regexp(t, "┃ 🗑️  subject1 will be deleted permanently\\W+Delete!\\W+Cancel.", render)
		})

		t.Run("Enter effectively deletes the subject", func(t *testing.T) {
			deleter.deletionResultMsg = DeletedSubjectMsg{"subject1", 1}

			subjectsPage.Update(keys.Key('d'))
			cmds := subjectsPage.Update(keys.Key(tea.KeyEnter))
			msgs := tests.ExecuteBatchCmd(cmds)

			assert.Contains(t, msgs, DeletedSubjectMsg{"subject1", 1})
		})

		t.Run("Display error when deletion fails", func(t *testing.T) {
			deleter.deletionResultMsg = sradmin.SubjectDeletionStartedMsg{}

			subjectsPage.Update(keys.Key(tea.KeyF2))
			subjectsPage.Update(keys.Key('d'))
			cmds := subjectsPage.Update(keys.Key(tea.KeyEnter))

			for _, msg := range tests.ExecuteBatchCmd(cmds) {
				subjectsPage.Update(msg)
			}
			subjectsPage.Update(sradmin.SubjectDeletionErrorMsg{
				Err: fmt.Errorf("unable to delete subject"),
			})

			render = subjectsPage.View(ui.NewTestKontext(), ui.TestRenderer)

			assert.Regexp(t, "unable to delete subject", render)

			t.Run("When deletion failure msg visible do allow other cmdbars to activate", func(t *testing.T) {
				subjectsPage.Update(keys.Key('/'))

				render = subjectsPage.View(ui.NewTestKontext(), ui.TestRenderer)

				assert.NotContains(t, render, "Failed to delete subject: unable to delete subject")
				assert.Contains(t, render, "> Search subject by name")
			})
		})

		t.Run("When deletion started show spinning indicator", func(t *testing.T) {

			subjectsPage, _ := New(
				&MockSubjectsLister{},
				&MockSubjectsDeleter{},
			)
			subjectsPage.Update(sradmin.SubjectDeletionStartedMsg{})

			render := subjectsPage.View(ui.TestKontext, ui.TestRenderer)

			assert.Contains(t, render, " ⏳ Deleting Subject")
		})
	})

	t.Run("When listing started show spinning indicator", func(t *testing.T) {

		subjectsPage, _ := New(
			&MockSubjectsLister{},
			&MockSubjectsDeleter{},
		)
		subjectsPage.Update(sradmin.SubjectListingStartedMsg{})

		render := subjectsPage.View(ui.TestKontext, ui.TestRenderer)

		assert.Contains(t, render, " ⏳ Loading subjects")
	})
}
