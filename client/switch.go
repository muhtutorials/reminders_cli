package client

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type idsFlag []string

func (ids *idsFlag) String() string {
	return strings.Join(*ids, ",")
}

func (ids *idsFlag) Set(v string) error {
	*ids = strings.Split(v, ",")
	return nil
}

type BackendHTTPClient interface {
	Create(title, message string, duration, retryPeriod time.Duration) ([]byte, error)
	Edit(id, title, message string, duration, retryPeriod time.Duration) ([]byte, error)
	Fetch(ids []string) ([]byte, error)
	Delete(ids []string) error
	Healthy(host string) bool
}

type Switch struct {
	client        BackendHTTPClient
	backendAPIURL string
	commands      map[string]func(string) error
}

func NewSwitch(url string) Switch {
	httpClient := NewHTTPClient(url)
	s := Switch{
		client:        httpClient,
		backendAPIURL: url,
	}
	s.commands = map[string]func(string) error{
		"create": s.create,
		"edit":   s.edit,
		"fetch":  s.fetch,
		"delete": s.delete,
		"health": s.health,
	}
	return s
}

func (s Switch) Switch() error {
	cmdName := os.Args[1]
	cmd, ok := s.commands[cmdName]
	if !ok {
		return fmt.Errorf("invalid command '%s'", cmdName)
	}
	return cmd(cmdName)
}

func (s Switch) Help() {
	var help string
	for name := range s.commands {
		help += name + "\t --help\n"
	}
	fmt.Printf("Usage of '%s':\n<command> [<args>]\n%s", os.Args[0], help)
}

func (s Switch) create(cmdName string) error {
	createCmd := flag.NewFlagSet(cmdName, flag.ExitOnError)
	title, message, duration, retryPeriod := s.reminderFlags(createCmd)

	if err := s.checkArgs(4); err != nil {
		return err
	}

	if err := s.parseCmd(createCmd); err != nil {
		return err
	}

	res, err := s.client.Create(*title, *message, *duration, *retryPeriod)
	if err != nil {
		return wrapError("could not create reminder", err)
	}

	fmt.Println("reminder created successfully:", string(res))
	return nil
}

func (s Switch) edit(cmdName string) error {
	ids := idsFlag{}
	editCmd := flag.NewFlagSet(cmdName, flag.ExitOnError)
	editCmd.Var(&ids, "id", "ID (int) of the reminder to edit")
	title, message, duration, retryPeriod := s.reminderFlags(editCmd)

	if err := s.checkArgs(2); err != nil {
		return err
	}

	if err := s.parseCmd(editCmd); err != nil {
		return err
	}

	lastID := ids[len(ids)-1]
	res, err := s.client.Edit(lastID, *title, *message, *duration, *retryPeriod)
	if err != nil {
		return wrapError("could not edit reminder", err)
	}

	fmt.Println("reminder edited successfully:", string(res))
	return nil
}

func (s Switch) fetch(cmdName string) error {
	ids := idsFlag{}
	fetchCmd := flag.NewFlagSet(cmdName, flag.ExitOnError)
	fetchCmd.Var(&ids, "id", "List of reminder IDs (int) to fetch")

	if err := s.checkArgs(1); err != nil {
		return err
	}

	if err := s.parseCmd(fetchCmd); err != nil {
		return err
	}

	res, err := s.client.Fetch(ids)
	if err != nil {
		return wrapError("could not fetch reminder(s)", err)
	}

	fmt.Println("reminder(s) fetched successfully:", string(res))
	return nil
}

func (s Switch) delete(cmdName string) error {
	ids := idsFlag{}
	deleteCmd := flag.NewFlagSet(cmdName, flag.ExitOnError)
	deleteCmd.Var(&ids, "id", "List of reminder IDs (int) to delete")

	if err := s.checkArgs(1); err != nil {
		return err
	}

	if err := s.parseCmd(deleteCmd); err != nil {
		return err
	}

	err := s.client.Delete(ids)
	if err != nil {
		return wrapError("could not delete reminder(s)", err)
	}

	fmt.Println("reminder(s) deleted successfully:", ids)
	return nil
}

func (s Switch) health(cmdName string) error {
	var host string
	healthCmd := flag.NewFlagSet(cmdName, flag.ExitOnError)
	healthCmd.StringVar(&host, "host", s.backendAPIURL, "Host to ping for health")

	if err := s.parseCmd(healthCmd); err != nil {
		return err
	}

	if s.client.Healthy(host) {
		fmt.Printf("Host %s is up and running", host)
	} else {
		fmt.Printf("Host %s is down", host)
	}

	return nil
}

func (s Switch) reminderFlags(f *flag.FlagSet) (*string, *string, *time.Duration, *time.Duration) {
	title, message, duration, retryPeriod := "", "", time.Duration(0), time.Duration(0)

	f.StringVar(&title, "title", "", "Reminder title")
	f.StringVar(&title, "t", "", "Reminder title")
	f.StringVar(&message, "message", "", "Reminder message")
	f.StringVar(&message, "m", "", "Reminder message")
	f.DurationVar(&duration, "duration", 0, "Reminder time")
	f.DurationVar(&duration, "d", 0, "Reminder time")
	f.DurationVar(&retryPeriod, "retry_period", 0, "Reminder retry period")
	f.DurationVar(&retryPeriod, "r", 0, "Reminder retry period")

	return &title, &message, &duration, &retryPeriod
}

func (s Switch) parseCmd(cmd *flag.FlagSet) error {
	err := cmd.Parse(os.Args[2:])
	if err != nil {
		return wrapError("could not parse '"+cmd.Name()+"' command flags", err)
	}
	return nil
}

func (s Switch) checkArgs(minArgs int) error {
	if len(os.Args) == 3 && os.Args[2] == "--help" {
		return nil
	}
	if len(os.Args)-2 < minArgs {
		fmt.Printf("incorrect use of %s\n%s %s --help\n", os.Args[1], os.Args[0], os.Args[1])
		return fmt.Errorf("%s expects at least %d arg(s), %d provided", os.Args[1], minArgs, len(os.Args)-2)
	}
	return nil
}
