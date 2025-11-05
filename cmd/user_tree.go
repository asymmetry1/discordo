package cmd


import (
    "github.com/ayn2op/discordo/internal/config"
    "github.com/ayn2op/discordo/internal/ui"
    "github.com/ayn2op/tview"
    "github.com/diamondburned/arikawa/v3/discord" 
)

type UserTree struct {
    *tview.List 
    cfg *config.Config
}

func newUserTree(cfg *config.Config) *UserTree {
    ut := &UserTree{
        List: tview.NewList(), 
        cfg:  cfg,
    }
  
    ut.Box = ui.ConfigureBox(ut.Box, &cfg.Theme) 

    ut.SetTitle("Members")

    return ut
}

func (ut *UserTree) Update(members []discord.Member) { 
    ut.Clear()
    for _, member := range members {
        ut.AddItem(member.User.DisplayOrUsername(), "", 0, nil)
    }
}

