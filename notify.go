package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"github.com/csotherden/strftime"
)

type Date string

const (
	today    Date = "Today"
	tomorrow Date = "Tomorrow"
)

type Event struct {
	Name      string
	Desc      string
	StartTime string
	Date      Date
}

func sameDateEST(t1, t2 time.Time) bool {
	t1 = t1.In(time.FixedZone("EST", -5*60*60))
	t2 = t2.In(time.FixedZone("EST", -5*60*60))
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func createSession() *discordgo.Session {
	if os.Getenv("BOT_TOKEN") == "" {
		fmt.Println("No Bot Token provided")
		return nil
	}
	dg, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		fmt.Printf("error creating Discord session: %v", err)
		return nil
	}
	err = dg.Open()
	if err != nil {
		fmt.Printf("error opening connection: %v", err)
		return nil
	}
	return dg
}

func getUpcomingCalendarEvents(session *discordgo.Session, guild string) []Event {
	events, err := session.GuildScheduledEvents(guild, false)
	if err != nil {
		fmt.Printf("error fetching scheduled events: %v", err)
		return nil
	}

	now := time.Now()
	var calendarEvents []Event

	// sort events by start time
	sort.Slice(events, func(i, j int) bool {
		return events[i].ScheduledStartTime.Before(events[j].ScheduledStartTime)
	})

	for _, e := range events {
		fmt.Println(e.Name, e.ScheduledStartTime)
		if sameDateEST(e.ScheduledStartTime, now) {

			calendarEvents = append(calendarEvents, Event{
				Name:      e.Name,
				Desc:      e.Description,
				StartTime: strftime.Format("%I:%M %p", e.ScheduledStartTime),
				Date:      today,
			})
		}
		if sameDateEST(e.ScheduledStartTime, now.Add(24*time.Hour)) {
			calendarEvents = append(calendarEvents, Event{
				Name:      e.Name,
				Desc:      e.Description,
				StartTime: strftime.Format("%I:%M %p", e.ScheduledStartTime),
				Date:      tomorrow,
			})
		}
	}
	return calendarEvents
}

func buildMessageEmbed(events []Event) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       "Upcoming Events",
		Description: "\u200b",
		Color:       0x00ff00, // Green color
		Fields:      []*discordgo.MessageEmbedField{},
	}

	eventFields := []*discordgo.MessageEmbedField{}
	for _, event := range events {
		field := &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s (%s, %s)", event.Name, event.Date, event.StartTime),
			Value:  event.Desc,
			Inline: false,
		}
		eventFields = append(eventFields, field)
	}

	embed.Fields = append(embed.Fields, eventFields...)

	return embed
}

func postEvents() {
	session := createSession()
	defer session.Close()

	guild := os.Getenv("GUILD_ID")
	channel := os.Getenv("CHANNEL_ID")
	if guild == "" || channel == "" {
		fmt.Println("GUILD_ID or CHANNEL_ID environment variable not set")
		return
	}

	events := getUpcomingCalendarEvents(session, guild)
	if len(events) == 0 {
		fmt.Println("No upcoming events")
		return
	}

	embed := buildMessageEmbed(events)
	_, err := session.ChannelMessageSendEmbed(channel, embed)
	if err != nil {
		fmt.Printf("error sending message: %v", err)
	}
}

func main() {
	lambda.Start(postEvents)
}
