package notifier

import (
	"fmt"
	"strings"
)

type ScalingMessage struct {
	Title string
	Body  string
}

func RenderScalingMessage(title string, event ScalingEvent, emoji string, level Level) *ScalingMessage {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%s **%s** %s\n\n", emoji, event.Decision, event.Fleet)
	fmt.Fprintf(&sb, "- **Fleet:** %s/%s\n", event.Namespace, event.Fleet)
	fmt.Fprintf(&sb, "- **Reason:** %s\n", event.Reason)
	fmt.Fprintf(&sb, "- **Replicas:** %d → %d\n", event.CurrentReplicas, event.DesiredReplicas)

	if len(event.Nodes) > 0 {
		fmt.Fprintf(&sb, "- **Nodes:** %s\n", strings.Join(event.Nodes, ", "))
	}

	fmt.Fprintf(&sb, "\n📅 %s", event.Timestamp.Format("2006-01-02 15:04:05"))

	return &ScalingMessage{
		Title: title,
		Body:  sb.String(),
	}
}
