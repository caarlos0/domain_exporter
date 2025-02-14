package whois

import (
    "fmt"
    "regexp"
    "time"
    "github.com/domainr/whois"
)

// kzAdapter implements custom adapter for .kz domains
type kzAdapter struct{}

func (a *kzAdapter) Prepare(req *whois.Request) error {
    return whois.DefaultAdapter.Prepare(req)
}

func (a *kzAdapter) Text(res *whois.Response) ([]byte, error) {
    text, err := whois.DefaultAdapter.Text(res)
    if err != nil {
        return nil, err
    }

    // Parse creation date from whois response
    createdStr := ""
    createdRegex := regexp.MustCompile(`Domain created: ([\d-]+ [\d:]+)`)
    if matches := createdRegex.FindSubmatch(text); len(matches) > 1 {
        createdStr = string(matches[1])
    }

    // Convert date to the format expected by exporter
    if createdStr != "" {
        created, err := time.Parse("2006-01-02 15:04:05", createdStr)
        if err != nil {
            return nil, fmt.Errorf("failed to parse creation date: %v", err)
        }
        
        // For .kz domains registration period is 1 year
        expiration := created.AddDate(1, 0, 0)
        
        // Add paid-till field in the same format as .ru domains
        response := string(text)
        response += fmt.Sprintf("\npaid-till: %s", expiration.Format("2006-01-02T15:04:05Z"))
        
        return []byte(response), nil
    }

    return text, nil
}

func init() {
    whois.BindAdapter(
        &kzAdapter{},
        "whois.nic.kz",
    )
}
