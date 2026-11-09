package notifier

import (
	"context"
	"fmt"
	"time"

	"finance-app-go-demo/pkg/logger"
	"finance-app-go-demo/pkg/mailer"
)

type Sender interface {
	Send(ctx context.Context, msg mailer.Message) error
}

func deliverAsync(ctx context.Context, sender Sender, log logger.Logger, msg mailer.Message, what string) {
	ctx = context.WithoutCancel(ctx)

	go func() {
		if err := sender.Send(ctx, msg); err != nil {
			log.ErrorContext(ctx, "failed to send "+what+" email", "email", msg.To, "error", err)
			return
		}

		log.InfoContext(ctx, what+" email sent", "email", msg.To)
	}()
}

func humanDuration(d time.Duration) string {
	d = d.Round(time.Minute)

	if d >= time.Hour && d%time.Hour == 0 {
		return plural(int(d/time.Hour), "hour")
	}

	return plural(int(d/time.Minute), "minute")
}

func plural(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, unit)
	}

	return fmt.Sprintf("%d %ss", n, unit)
}
