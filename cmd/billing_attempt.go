package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func billingAttemptCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "billing-attempt",
		Aliases: []string{"ba"},
		Short:   "Manage billing attempts",
	}

	cmd.AddCommand(baRescheduleCmd(profile, jsonOut, agentOut))
	cmd.AddCommand(baDeleteCmd(profile, jsonOut, agentOut))
	cmd.AddCommand(baSkipCmd(profile, jsonOut, agentOut))
	cmd.AddCommand(baUnskipCmd(profile, jsonOut, agentOut))

	return cmd
}

// ─── Reschedule ──────────────────────────────────────────────────────────────

func baRescheduleCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
	var (
		id             int
		subscriptionID int
		date           string
		timeStr        string
		timezone       string
		resetSchedule  bool
	)

	cmd := &cobra.Command{
		Use:     "reschedule",
		Aliases: []string{"rs"},
		Short:   "Reschedule a billing attempt",
		Example: `  seal-cli billing-attempt reschedule --id 123 --subscription-id 456 \
    --date 2025-12-01 --time 14:30 --timezone "-05:00"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == 0 {
				return fmt.Errorf("--id is required")
			}
			if subscriptionID == 0 {
				return fmt.Errorf("--subscription-id is required")
			}
			if date == "" {
				return fmt.Errorf("--date is required (format: YYYY-MM-DD)")
			}
			if timeStr == "" {
				return fmt.Errorf("--time is required (format: HH:MM)")
			}
			if timezone == "" {
				return fmt.Errorf("--timezone is required (e.g. \"+00:00\")")
			}

			client, err := newClient(*profile)
			if err != nil {
				return err
			}
			data, err := client.RescheduleBillingAttempt(id, subscriptionID, date, timeStr, timezone, resetSchedule)
			if err != nil {
				return err
			}
			return renderSuccess(data, *jsonOut, *agentOut, "reschedule", id, fmt.Sprintf("Billing attempt #%d rescheduled to %s %s %s.", id, date, timeStr, timezone))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Billing attempt ID")
	cmd.Flags().IntVar(&subscriptionID, "subscription-id", 0, "Subscription ID")
	cmd.Flags().StringVar(&date, "date", "", "New date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&timeStr, "time", "", "New time (HH:MM)")
	cmd.Flags().StringVar(&timezone, "timezone", "", "Timezone offset (e.g. +00:00 or -05:00)")
	cmd.Flags().BoolVar(&resetSchedule, "reset-schedule", false, "Reset rest of the schedule from this date")

	return cmd
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func baDeleteCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
	var (
		id             int
		subscriptionID int
	)

	cmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"rm", "d"},
		Short:   "Delete a billing attempt",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == 0 {
				return fmt.Errorf("--id is required")
			}
			if subscriptionID == 0 {
				return fmt.Errorf("--subscription-id is required")
			}

			client, err := newClient(*profile)
			if err != nil {
				return err
			}
			data, err := client.DeleteBillingAttempt(id, subscriptionID)
			if err != nil {
				return err
			}
			return renderSuccess(data, *jsonOut, *agentOut, "delete", id, fmt.Sprintf("Billing attempt #%d deleted.", id))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Billing attempt ID")
	cmd.Flags().IntVar(&subscriptionID, "subscription-id", 0, "Subscription ID")

	return cmd
}

// ─── Skip ────────────────────────────────────────────────────────────────────

func baSkipCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
	var (
		id             int
		subscriptionID int
	)

	cmd := &cobra.Command{
		Use:     "skip",
		Aliases: []string{"sk"},
		Short:   "Skip a billing attempt",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == 0 {
				return fmt.Errorf("--id is required")
			}
			if subscriptionID == 0 {
				return fmt.Errorf("--subscription-id is required")
			}

			client, err := newClient(*profile)
			if err != nil {
				return err
			}
			data, err := client.SkipBillingAttempt(id, subscriptionID)
			if err != nil {
				return err
			}
			return renderSuccess(data, *jsonOut, *agentOut, "skip", id, fmt.Sprintf("Billing attempt #%d skipped.", id))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Billing attempt ID")
	cmd.Flags().IntVar(&subscriptionID, "subscription-id", 0, "Subscription ID")

	return cmd
}

// ─── Unskip ──────────────────────────────────────────────────────────────────

func baUnskipCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
	var (
		id             int
		subscriptionID int
	)

	cmd := &cobra.Command{
		Use:     "unskip",
		Aliases: []string{"us"},
		Short:   "Unskip a previously skipped billing attempt",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == 0 {
				return fmt.Errorf("--id is required")
			}
			if subscriptionID == 0 {
				return fmt.Errorf("--subscription-id is required")
			}

			client, err := newClient(*profile)
			if err != nil {
				return err
			}
			data, err := client.UnskipBillingAttempt(id, subscriptionID)
			if err != nil {
				return err
			}
			return renderSuccess(data, *jsonOut, *agentOut, "unskip", id, fmt.Sprintf("Billing attempt #%d unskipped.", id))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Billing attempt ID")
	cmd.Flags().IntVar(&subscriptionID, "subscription-id", 0, "Subscription ID")

	return cmd
}
