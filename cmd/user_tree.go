package cmd


import (
		"log/slog"
		"sort"
		"fmt"

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
    cfg:  cfg,
  }
  
  ut.Box = ui.ConfigureBox(ut.Box, &cfg.Theme) 

  ut.
		SetRoot(tview.NewTreeNode("")).
		SetTopLevel(1).
		SetTitle("Members")
		//planning on select func -> member modal
		//theme soon

    return ut
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
	
	// Tree Node
	for _, role := range roles {
		if role.ID == discord.RoleID(guildID) || !role.Hoist {
			continue
		}

		var roleMembers []discord.Member

		for _, member := range members {
			if memberHasRole(member, role.ID) {
				roleMembers = append(roleMembers, member)
			}
		}

		if len(roleMembers) == 0 {
			continue
		}

		roleNode := tview.NewTreeNode(fmt.Sprintf("%s (%d)", role.Name, len(roleMembers))).
			SetColor(tcell.GetColor(role.Color.String()))
		root.AddChild(roleNode)

		for _, member := range roleMembers {
			memberNode := tview.NewTreeNode(member.User.DisplayOrUsername())
			roleNode.AddChild(memberNode)
		}
	}
	
	var unclassifiedMembers []discord.Member
	for _, member := range members {
		isOnlyEveryone := len(member.RoleIDs) == 1 && member.RoleIDs[0] == discord.RoleID(guildID)
		isNoRoles := len(member.RoleIDs) == 0

		if isOnlyEveryone || isNoRoles {
			unclassifiedMembers = append(unclassifiedMembers, member)
		}
	}
	
	if len(unclassifiedMembers) > 0 {
		otherNode := tview.NewTreeNode(fmt.Sprintf("Online (%d)", len(unclassifiedMembers)))
		root.AddChild(otherNode)
		
		for _, member := range unclassifiedMembers {
			memberNode := tview.NewTreeNode(member.User.DisplayOrUsername())
			otherNode.AddChild(memberNode)
		}
	}

	root.Walk(func(node, parent *tview.TreeNode) bool {
		node.SetExpanded(true)
		return true
	})
}

//helper
func memberHasRole(member discord.Member, roleID discord.RoleID) bool {
	for _, id := range member.RoleIDs {
		if id == roleID {
			return true
		}
	}
	return false
}

/* WIP
func getStatusPrefix(userID discord.UserID) string {
	presence, err :- discordState.Cabinet.Presence(userID)
	if err != nil {
		return ""
	}

	switch presence.Status {
	case discord.Online:
		return "[green]"
	}
}
*/
