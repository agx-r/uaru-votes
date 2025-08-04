package handlers

import (
	"fmt"
	"time"

	"github.com/uaru-shit/votes/internal/domain"
	"github.com/uaru-shit/votes/pkg/utils"
	tb "gopkg.in/telebot.v4"
)

func handlePollResults(ctx domain.Context, msg *tb.Message, member *tb.ChatMember) {
	log := ctx.Log()
	bot := ctx.Bot()
	time.Sleep(time.Hour)

	poll, err := bot.StopPoll(msg)
	if err != nil {
		log.Error("failed to stop the poll", utils.ErrorAttr(err))

		return
	}

	shouldMute := poll.Options[0].VoterCount > poll.Options[1].VoterCount

	if _, err := bot.Reply(msg, map[bool]string{
		true:  "Мут nyuuu",
		false: "Размучен.",
	}[shouldMute]); err != nil {
		log.Error("failed to reply to poll", utils.ErrorAttr(err))
	}

	perm := !shouldMute // value of permissions is opposite of shouldMute

	member.RestrictedUntil = map[bool]int64{
		true:  tb.Forever(), // tb.Forever() is int64
		false: 0,
	}[shouldMute]
	member.CanSendMessages = perm
	member.CanSendMedia = perm
	member.CanSendPolls = perm
	member.CanSendOther = perm
	member.CanAddPreviews = perm

	if err := bot.Restrict(ctx.Chat(), member); err != nil {
		logWord := map[bool]string{true: "mute", false: "unmute"}[shouldMute]
		msgWord := map[bool]string{true: "замутить", false: "размутить"}[shouldMute]

		log.Error(fmt.Sprintf("cannot %s user", logWord), utils.ErrorAttr(err))

		if _, err := bot.Reply(msg, "Чота не могу "+msgWord); err != nil {
			log.Error("can't even cry", utils.ErrorAttr(err))
		}
	}
}

func HandleVoteban(ctx domain.Context) error {
	bot := ctx.Bot()

	if !ctx.Message().FromGroup() {
		return ctx.Reply("В лс не баню сори")
	}

	if ctx.Message().ReplyTo == nil {
		return ctx.Reply("Ответь на сообщение кого забанить")
	}

	userToBan := ctx.Message().ReplyTo.Sender
	member, err := bot.ChatMemberOf(ctx.Chat(), userToBan)

	if err != nil {
		return fmt.Errorf("failed to get member: %w", err)
	}

	admins, err := bot.AdminsOf(ctx.Chat())
	if err != nil {
		return fmt.Errorf("failed to get admins: %w", err)
	}

	if utils.IsAdmin(userToBan.ID, admins) {
		return ctx.Reply("ммм не")
	}

	if !utils.BotCanMute(ctx.BotUser().ID, admins) {
		return ctx.Reply("Админом меня сделай, олух")
	}

	msg, err := bot.Send(ctx.Chat(), &tb.Poll{
		Question:  "Забанить или разбанить?",
		Anonymous: false,
		Options: []tb.PollOption{
			{Text: "Мут"},
			{Text: "Размут"},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send poll: %w", err)
	}

	go handlePollResults(ctx, msg, member)

	return nil
}
