package mail

import "context"

type Sender struct{ client *ZeptoClient }

func NewSender(client *ZeptoClient) *Sender { return &Sender{client: client} }

func (s *Sender) Send(ctx context.Context, to, subject, html string) error {
	return s.client.Send(ctx, Message{To: to, Subject: subject, HTML: html})
}
