package clusters_tab

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/styles"
	"ktea/tests/keys"
	"ktea/ui"
	"ktea/ui/pages/clusters_page"
	"testing"
)

type MockConnectionCheckStartedMsg struct {
	cluster *config.Cluster
}

func mockConnChecker(cluster *config.Cluster) tea.Msg {
	return MockConnectionCheckStartedMsg{cluster: cluster}
}

func TestClustersTab(t *testing.T) {
	var ktx = kontext.ProgramKtx{
		Config: &config.Config{
			Clusters: []config.Cluster{},
		},
		WindowWidth:  100,
		WindowHeight: 100,
	}
	t.Run("When no cluster exists, open create new form", func(t *testing.T) {
		// given
		clustersTab, _ := New(&kontext.ProgramKtx{
			Config:       &config.Config{},
			WindowWidth:  0,
			WindowHeight: 0,
		}, mockConnChecker)

		// when
		render := clustersTab.View(&ktx, ui.TestRenderer)

		// then
		assert.Contains(t, render, "┃ Name")
	})

	t.Run("List clusters page", func(t *testing.T) {
		t.Run("opens when at least one env exists", func(t *testing.T) {
			// given
			programKtx := &kontext.ProgramKtx{
				Config: &config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "prd",
							Color:            "#808080",
							Active:           true,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
					},
				},
				WindowWidth:  100,
				WindowHeight: 100,
			}
			var clustersTab, _ = New(programKtx, mockConnChecker)

			// when
			render := clustersTab.View(programKtx, ui.TestRenderer)

			// then
			assert.Contains(t, render, "prd")
		})

		t.Run("indicates active cluster in list", func(t *testing.T) {
			// given
			programKtx := &kontext.ProgramKtx{
				Config: &config.Config{
					Clusters: []config.Cluster{
						{
							Name:             "prd",
							Color:            styles.ColorRed,
							Active:           false,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
						{
							Name:             "tst",
							Color:            styles.ColorGreen,
							Active:           true,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
						{
							Name:             "uat",
							Color:            styles.ColorBlue,
							Active:           false,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
					},
				},
				WindowWidth:     100,
				WindowHeight:    100,
				AvailableHeight: 100,
			}
			var clustersTab, _ = New(programKtx, mockConnChecker)

			// when
			render := clustersTab.View(programKtx, ui.TestRenderer)

			// then
			assert.Regexp(t, "X\\W+tst", render)
			assert.Regexp(t, "│\\W+prd", render)
			assert.Regexp(t, "│\\W+uat", render)
			assert.NotRegexp(t, "X\\W+prd", render)
			assert.NotRegexp(t, "X\\W+uat", render)
		})

		t.Run("enter starts a connectivity check for the selected cluster", func(t *testing.T) {
			// given
			programKtx := &kontext.ProgramKtx{
				Config: &config.Config{
					ConfigIO: &config.InMemoryConfigIO{},
					Clusters: []config.Cluster{
						{
							Name:             "prd",
							Color:            styles.ColorRed,
							Active:           true,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
						{
							Name:             "tst",
							Color:            styles.ColorGreen,
							Active:           false,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
						{
							Name:             "uat",
							Color:            styles.ColorBlue,
							Active:           false,
							BootstrapServers: []string{"localhost:9092"},
							SASLConfig:       nil,
						},
					},
				},
				WindowWidth:     100,
				WindowHeight:    100,
				AvailableHeight: 100,
			}
			var clustersTab, _ = New(programKtx, mockConnChecker)
			// and table has been initialized
			clustersTab.View(programKtx, ui.TestRenderer)

			// when
			clustersTab.Update(keys.Key(tea.KeyDown))
			cmds := clustersTab.Update(keys.Key(tea.KeyEnter))
			msgs := executeBatchCmd(cmds)
			assert.NotNil(t, msgs)

			// then the connectivity check has been started
			assert.IsType(t, MockConnectionCheckStartedMsg{}, msgs[0])

			t.Run("and shows a spinner", func(t *testing.T) {
				clustersTab.Update(kadmin.ConnCheckStartedMsg{
					Cluster: &config.Cluster{
						Name: "tst",
					},
				})

				rendered := clustersTab.View(programKtx, ui.TestRenderer)

				assert.Contains(t, rendered, "Checking connectivity to tst")
			})

			t.Run("and shows failure msg upon connectivity error", func(t *testing.T) {
				clustersTab.Update(kadmin.ConnCheckErrMsg{
					Err: fmt.Errorf("kafka: client has run out of available brokers to talk to: EOF"),
				})

				rendered := clustersTab.View(programKtx, ui.TestRenderer)

				assert.Contains(t, rendered, "Connection check failed: kafka: client has run out of available brokers to talk to: EOF")
			})

			t.Run("and shows success msg upon connection ", func(t *testing.T) {
				clustersTab.Update(kadmin.ConnCheckSucceededMsg{})

				rendered := clustersTab.View(programKtx, ui.TestRenderer)

				assert.Contains(t, rendered, "Connection check succeeded, switching cluster")
			})

			t.Run("and activated is indicated", func(t *testing.T) {
				programKtx.Config.SwitchCluster("tst")
				clustersTab.Update(clusters_page.ClusterSwitchedMsg{
					Cluster: &config.Cluster{
						Name: "tst",
					},
				})

				rendered := clustersTab.View(programKtx, ui.TestRenderer)

				assert.Regexp(t, "X\\W+tst", rendered)
				assert.Regexp(t, "│\\W+prd", rendered)
				assert.Regexp(t, "│\\W+uat", rendered)
				assert.NotRegexp(t, "X\\W+prd", rendered)
			})

			t.Run("Activated cluster is selected", func(t *testing.T) {
				assert.Equal(t, "tst",
					*(clustersTab.active.(*clusters_page.Model)).SelectedCluster())
			})
		})
	})

	t.Run("Delete cluster", func(t *testing.T) {
		cfg := config.New(&config.InMemoryConfigIO{})
		cfg.RegisterCluster(config.RegistrationDetails{
			Name:       "prd",
			Color:      styles.ColorRed,
			Host:       "localhost:9092",
			AuthMethod: config.NoneAuthMethod,
		})
		cfg.RegisterCluster(config.RegistrationDetails{
			Name:       "tst",
			Color:      styles.ColorRed,
			Host:       "localhost:9093",
			AuthMethod: config.NoneAuthMethod,
		})
		cfg.RegisterCluster(config.RegistrationDetails{
			Name:       "uat",
			Color:      styles.ColorRed,
			Host:       "localhost:9093",
			AuthMethod: config.NoneAuthMethod,
		})
		programKtx := &kontext.ProgramKtx{
			Config:          cfg,
			WindowWidth:     100,
			WindowHeight:    100,
			AvailableHeight: 100,
		}

		t.Run("F2 raises delete confirmation", func(t *testing.T) {
			// given
			var clustersTab, _ = New(programKtx, mockConnChecker)
			// and table has been initialized
			render := clustersTab.View(programKtx, ui.TestRenderer)

			// when
			clustersTab.Update(keys.Key(tea.KeyF2))

			// then
			render = clustersTab.View(programKtx, ui.TestRenderer)
			assert.Contains(t, render, "🗑️  prd will be deleted permanently")
		})

		t.Run("esc cancels raised delete confirmation", func(t *testing.T) {
			// given
			var clustersTab, _ = New(programKtx, mockConnChecker)
			// and table has been initialized
			render := clustersTab.View(programKtx, ui.TestRenderer)
			// and delete confirmation has been raised
			clustersTab.Update(keys.Key(tea.KeyF2))
			render = clustersTab.View(programKtx, ui.TestRenderer)
			assert.Contains(t, render, "🗑️  prd will be deleted permanently")

			// when
			clustersTab.Update(keys.Key(tea.KeyEsc))

			// then
			render = clustersTab.View(programKtx, ui.TestRenderer)
			assert.NotContains(t, render, "🗑️  prd will be deleted permanently")
		})

		t.Run("enter deletes cluster", func(t *testing.T) {
			// given
			var clustersTab, _ = New(programKtx, mockConnChecker)
			// and table has been initialized
			render := clustersTab.View(programKtx, ui.TestRenderer)

			// when
			clustersTab.Update(keys.Key(tea.KeyDown))
			clustersTab.Update(keys.Key(tea.KeyF2))
			clustersTab.Update(keys.Key('d'))
			cmds := clustersTab.Update(keys.Key(tea.KeyEnter))
			msgs := executeBatchCmd(cmds)
			for _, msg := range msgs {
				clustersTab.Update(msg)
			}

			// then
			render = clustersTab.View(programKtx, ui.TestRenderer)
			assert.NotContains(t, render, "tst")
		})

		t.Run("unable to delete active cluster", func(t *testing.T) {
			// given
			cfg := config.New(&config.InMemoryConfigIO{})
			cfg.RegisterCluster(config.RegistrationDetails{
				Name:       "prd",
				Color:      styles.ColorRed,
				Host:       "localhost:9092",
				AuthMethod: config.NoneAuthMethod,
			})
			cfg.RegisterCluster(config.RegistrationDetails{
				Name:       "tst",
				Color:      styles.ColorRed,
				Host:       "localhost:9093",
				AuthMethod: config.NoneAuthMethod,
			})
			cfg.RegisterCluster(config.RegistrationDetails{
				Name:       "uat",
				Color:      styles.ColorRed,
				Host:       "localhost:9093",
				AuthMethod: config.NoneAuthMethod,
			})
			programKtx := &kontext.ProgramKtx{
				Config:          cfg,
				WindowWidth:     100,
				WindowHeight:    100,
				AvailableHeight: 100,
			}
			// and
			var clustersTab, _ = New(programKtx, mockConnChecker)
			// and table has been initialized
			render := clustersTab.View(programKtx, ui.TestRenderer)

			// when
			clustersTab.Update(keys.Key(tea.KeyF2))
			clustersTab.Update(keys.Key('d'))
			cmds := clustersTab.Update(keys.Key(tea.KeyEnter))
			msgs := executeBatchCmd(cmds)
			for _, msg := range msgs {
				clustersTab.Update(msg)
			}

			// then
			render = clustersTab.View(programKtx, ui.TestRenderer)
			assert.Contains(t, render, "Unable to delete: active cluster")
		})
	})

	t.Run("Edit cluster", func(t *testing.T) {
		cfg := config.New(&config.InMemoryConfigIO{})
		cfg.RegisterCluster(config.RegistrationDetails{
			Name:       "prd",
			Color:      styles.ColorRed,
			Host:       "localhost:9092",
			AuthMethod: config.NoneAuthMethod,
		})
		cfg.RegisterCluster(config.RegistrationDetails{
			Name:       "tst",
			Color:      styles.ColorRed,
			Host:       "localhost:9093",
			AuthMethod: config.NoneAuthMethod,
		})
		cfg.RegisterCluster(config.RegistrationDetails{
			Name:       "uat",
			Color:      styles.ColorRed,
			Host:       "localhost:9093",
			AuthMethod: config.NoneAuthMethod,
		})
		programKtx := &kontext.ProgramKtx{
			Config:       cfg,
			WindowWidth:  100,
			WindowHeight: 100,
		}

		t.Run("c-e opens edit page", func(t *testing.T) {
			// given
			var clustersTab, _ = New(programKtx, mockConnChecker)
			// and table has been initialized
			clustersTab.View(programKtx, ui.TestRenderer)

			// when
			clustersTab.Update(keys.Key(tea.KeyCtrlE))

			// then
			render := clustersTab.View(programKtx, ui.TestRenderer)
			assert.Contains(t, render, "> prd")
			assert.Contains(t, render, "> localhost:9092")
		})
	})
}

func executeBatchCmd(cmd tea.Cmd) []tea.Msg {
	var msgs []tea.Msg
	if cmd == nil {
		return msgs
	}

	msg := cmd()
	if msg == nil {
		return msgs
	}

	// If the message is a batch, process its commands
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, subCmd := range batch {
			if subCmd != nil {
				msgs = append(msgs, executeBatchCmd(subCmd)...)
			}
		}
		return msgs
	}

	// Otherwise, it's a normal message
	msgs = append(msgs, msg)
	return msgs
}
