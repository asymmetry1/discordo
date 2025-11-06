package cmd


import (
	"fmt"
	"log/slog"
	"sort"

	"github.com/ayn2op/discordo/internal/config"
	"github.com/ayn2op/discordo/internal/ui"
	"github.com/ayn2op/tview"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/gdamore/tcell/v2"
)

type UserTree struct {
	*tview.TreeView
	cfg *config.Config
}

func newUserTree(cfg *config.Config) *UserTree {
	ut := &UserTree{
		TreeView: tview.NewTreeView(),
		cfg:      cfg,
	}

	ut.Box = ui.ConfigureBox(ut.Box, &cfg.Theme)
	ut.SetRoot(tview.NewTreeNode("")).
		SetTopLevel(1).
		SetTitle("Members")

	return ut
}

func getStatus(guildID discord.GuildID, userID discord.UserID) (string, bool) {
	presence, err := discordState.Cabinet.Presence(guildID, userID)
	if err != nil {
		// Missing presence = offline
		return "[gray]○ [white]", true
	}

	switch presence.Status {
	case discord.OnlineStatus:
		return "[green]● [white]", false
	case "idle":
		return "[yellow]● [white]", false
	case "dnd":
		return "[red]● [white]", false
	case "offline", "invisible":
		return "[gray]○ [white]", true
	default:
		return "[gray]○ [white]", true
	}
}

func (ut *UserTree) Update(guildID discord.GuildID, members []discord.Member) {
	root := ut.GetRoot()
	root.ClearChildren()

	roles, err := discordState.Cabinet.Roles(guildID)
	if err != nil {
		slog.Error("failed to get guild roles", "err", err)
		return
	}

	sort.Slice(roles, func(i, j int) bool {
		if roles[i].Position != roles[j].Position {
			return roles[i].Position > roles[j].Position
		}
		return roles[i].ID < roles[j].ID
	})

	// Root sections
	onlineNode := tview.NewTreeNode("Online")
	offlineNode := tview.NewTreeNode("Offline")
	root.AddChild(onlineNode)
	root.AddChild(offlineNode)

	classified := make(map[discord.UserID]bool)

	// ONLINE by roles
	for _, role := range roles {
		if role.ID == discord.RoleID(guildID) || !role.Hoist {
			continue
		}

		var roleOnline []*tview.TreeNode
		for _, m := range members {
			if classified[m.User.ID] {
				continue
			}

			prefix, offline := getStatus(guildID, m.User.ID)
			if offline {
				continue
			}

			if memberHasRole(m, role.ID) {
				name := m.User.DisplayOrUsername()
				memberNode := tview.NewTreeNode(prefix + name).
					SetColor(tcell.GetColor(role.Color.String()))
				roleOnline = append(roleOnline, memberNode)
				classified[m.User.ID] = true
			}
		}

		if len(roleOnline) > 0 {
			roleNode := tview.NewTreeNode(fmt.Sprintf("%s (%d)", role.Name, len(roleOnline))).
				SetColor(tcell.GetColor(role.Color.String()))
			for _, node := range roleOnline {
				roleNode.AddChild(node)
			}
			onlineNode.AddChild(roleNode)
		}
	}

	// ONLINE without roles
	for _, m := range members {
		if classified[m.User.ID] {
			continue
		}

		prefix, offline := getStatus(guildID, m.User.ID)
		if offline {
			continue
		}

		name := m.User.DisplayOrUsername()
		onlineNode.AddChild(tview.NewTreeNode(prefix + name))
		classified[m.User.ID] = true
	}

	// OFFLINE (everyone else)
	for _, m := range members {
		if classified[m.User.ID] {
			continue
		}

		prefix, offline := getStatus(guildID, m.User.ID)
		if !offline {
			continue
		}

		name := m.User.DisplayOrUsername()
		offlineNode.AddChild(tview.NewTreeNode(prefix + name))
		classified[m.User.ID] = true
	}

	root.Walk(func(node, parent *tview.TreeNode) bool {
		node.SetExpanded(true)
		return true
	})
}

// helper
func memberHasRole(member discord.Member, roleID discord.RoleID) bool {
	for _, id := range member.RoleIDs {
		if id == roleID {
			return true
		}
	}
	return false
}
