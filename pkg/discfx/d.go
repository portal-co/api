package discfx

import (
	"context"
	"crypto/sha256"
	"encoding/base64"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
)

func Client(key string) fx.Option {
	return fx.Provide(func(l fx.Lifecycle) (*discordgo.Session, error) {
		s, err := discordgo.New(key)
		if err != nil {
			return nil, err
		}
		s.Identify.Intents = discordgo.IntentMessageContent
		l.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return s.Open()
			},
			OnStop: func(ctx context.Context) error {
				return s.Close()
			},
		})
		return s, nil
	})
}
func Uhash(u *discordgo.User, token []byte) string {
	s := sha256.New()
	s.Write(token)
	s.Write([]byte(u.ID))
	return base64.StdEncoding.EncodeToString(s.Sum(token))
}
func AddCmd(s *discordgo.Session, l fx.Lifecycle, v *discordgo.ApplicationCommand, a func(s *discordgo.Session, i *discordgo.InteractionCreate)) {
	var revert func()
	var cmd *discordgo.ApplicationCommand
	l.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			var err error
			cmd, err = s.ApplicationCommandCreate(s.State.User.ID, "", v)
			if err != nil {
				return err
			}
			revert = s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
				if i.Data.Type() == discordgo.InteractionApplicationCommand {
					if i.ApplicationCommandData().ID == cmd.ID {
						a(s, i)
					}
				}
			})
			return nil
		},
		OnStop: func(ctx context.Context) error {
			revert()
			revert = nil
			i := cmd.ID
			cmd = nil
			return s.ApplicationCommandDelete(s.State.User.ID, "", i)
		},
	})
}
