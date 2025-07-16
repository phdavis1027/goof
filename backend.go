package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

func SearchErrorReports(filter Filter) ([]ErrorReport, error) {
	config := LoadConfig()
	logToFile("DEBUG: SearchErrorReports - Creating Meilisearch client with URL: %s, Key: '%s' (len=%d)\n",
		config.MeilisearchURL, config.MeilisearchKey, len(config.MeilisearchKey))

	client := meilisearch.New(config.MeilisearchURL, meilisearch.WithAPIKey(config.MeilisearchKey))
	index := client.Index(config.IndexName)

	// Build search query for full-text search
	var queryParts []string

	// Add general query if provided
	if filter.Q != "" {
		queryParts = append(queryParts, filter.Q)
	}

	// Add specific field searches to the query
	if filter.Symptom != "" {
		queryParts = append(queryParts, filter.Symptom)
	}
	if filter.Program != "" {
		queryParts = append(queryParts, filter.Program)
	}
	if filter.ProgramVersion != "" {
		queryParts = append(queryParts, filter.ProgramVersion)
	}
	if filter.Distro != "" {
		queryParts = append(queryParts, filter.Distro)
	}
	if filter.DistroVersion != "" {
		queryParts = append(queryParts, filter.DistroVersion)
	}
	if filter.Solution != "" {
		queryParts = append(queryParts, filter.Solution)
	}

	// Combine all query parts
	searchQuery := strings.Join(queryParts, " ")

	// Build filter expressions (only for non-text fields like dates and exact matches)
	var filters []string

	if filter.DateFrom != nil {
		filters = append(filters, fmt.Sprintf("date >= %d", filter.DateFrom.Unix()))
	}
	if filter.DateTo != nil {
		filters = append(filters, fmt.Sprintf("date <= %d", filter.DateTo.Unix()))
	}
	if len(filter.ResourcesAny) > 0 {
		resourceFilters := make([]string, len(filter.ResourcesAny))
		for i, resource := range filter.ResourcesAny {
			resourceFilters[i] = fmt.Sprintf("resources = \"%s\"", resource)
		}
		filters = append(filters, fmt.Sprintf("(%s)", strings.Join(resourceFilters, " OR ")))
	}

	searchRequest := &meilisearch.SearchRequest{
		Limit: 100,
	}

	if len(filters) > 0 {
		searchRequest.Filter = strings.Join(filters, " AND ")
	}

	searchResponse, err := index.Search(searchQuery, searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	var reports []ErrorReport
	for _, hit := range searchResponse.Hits {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		report := ErrorReport{
			ID:             getString(hitMap, "id"),
			Symptom:        getString(hitMap, "symptom"),
			Program:        getString(hitMap, "program"),
			ProgramVersion: getString(hitMap, "program_version"),
			Distro:         getString(hitMap, "distro"),
			DistroVersion:  getString(hitMap, "distro_version"),
			Solution:       getString(hitMap, "solution"),
			Resources:      getStringArray(hitMap, "resources"),
		}

		// Convert Unix timestamp back to time.Time
		if dateField, ok := hitMap["date"]; ok {
			if dateFloat, ok := dateField.(float64); ok {
				report.Date = time.Unix(int64(dateFloat), 0)
			}
		}

		reports = append(reports, report)
	}

	return reports, nil
}

func SaveErrorReport(report ErrorReport) error {
	config := LoadConfig()
	logToFile("DEBUG: SaveErrorReport - Creating Meilisearch client with URL: %s, Key: '%s' (len=%d)\n",
		config.MeilisearchURL, config.MeilisearchKey, len(config.MeilisearchKey))

	logToFile("%+v\n", report)

	client := meilisearch.New(config.MeilisearchURL, meilisearch.WithAPIKey(config.MeilisearchKey))
	index := client.Index(config.IndexName)

	// Generate unique ID based on timestamp and program
	id := fmt.Sprintf("%d-%s", time.Now().UnixNano(), report.Program)

	// Create document with ID
	document := map[string]interface{}{
		"id":              id,
		"symptom":         report.Symptom,
		"date":            report.Date.Unix(), // Store as Unix timestamp for filtering
		"program":         report.Program,
		"program_version": report.ProgramVersion,
		"distro":          report.Distro,
		"distro_version":  report.DistroVersion,
		"resources":       report.Resources,
		"solution":        report.Solution,
	}

	_, err := index.AddDocuments([]map[string]interface{}{document})
	if err != nil {
		return fmt.Errorf("failed to save error report: %w", err)
	}

	return nil
}

func UpdateErrorReport(report ErrorReport, originalID string) error {
	config := LoadConfig()
	logToFile("DEBUG: UpdateErrorReport - Creating Meilisearch client with URL: %s, Key: '%s' (len=%d)\n",
		config.MeilisearchURL, config.MeilisearchKey, len(config.MeilisearchKey))

	logToFile("Updating report with ID: %s, %+v\n", originalID, report)

	client := meilisearch.New(config.MeilisearchURL, meilisearch.WithAPIKey(config.MeilisearchKey))
	index := client.Index(config.IndexName)

	// Create updated document with same ID
	document := map[string]interface{}{
		"id":              originalID,
		"symptom":         report.Symptom,
		"date":            report.Date.Unix(), // Store as Unix timestamp for filtering
		"program":         report.Program,
		"program_version": report.ProgramVersion,
		"distro":          report.Distro,
		"distro_version":  report.DistroVersion,
		"resources":       report.Resources,
		"solution":        report.Solution,
	}

	// Update the document (Meilisearch will replace the existing document with the same ID)
	_, err := index.AddDocuments([]map[string]interface{}{document})
	if err != nil {
		return fmt.Errorf("failed to update error report: %w", err)
	}

	return nil
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getStringArray(m map[string]interface{}, key string) []string {
	if val, ok := m[key]; ok {
		if arr, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return []string{}
}

func DeleteErrorReport(id string) error {
	config := LoadConfig()
	logToFile("DEBUG: DeleteErrorReport - Creating Meilisearch client with URL: %s, Key: '%s' (len=%d)\n",
		config.MeilisearchURL, config.MeilisearchKey, len(config.MeilisearchKey))

	logToFile("Deleting report with ID: %s\n", id)

	client := meilisearch.New(config.MeilisearchURL, meilisearch.WithAPIKey(config.MeilisearchKey))
	index := client.Index(config.IndexName)

	// Delete the document from Meilisearch
	_, err := index.DeleteDocument(id)
	if err != nil {
		return fmt.Errorf("failed to delete error report: %w", err)
	}

	return nil
}

func InitializeIndexIfNeeded() error {
	config := LoadConfig()
	client := meilisearch.New(config.MeilisearchURL, meilisearch.WithAPIKey(config.MeilisearchKey))
	index := client.Index(config.IndexName)

	// Define searchable attributes for full-text search
	searchableAttributes := []string{
		"symptom",
		"program",
		"program_version",
		"distro",
		"distro_version",
		"solution",
	}

	// Define filterable attributes for exact filtering (dates, resources)
	filterableAttributes := []string{
		"date",
		"resources",
	}

	// Update searchable attributes
	_, err := index.UpdateSearchableAttributes(&searchableAttributes)
	if err != nil {
		return fmt.Errorf("failed to update searchable attributes: %w", err)
	}

	// Update filterable attributes
	_, err = index.UpdateFilterableAttributes(&filterableAttributes)
	if err != nil {
		return fmt.Errorf("failed to update filterable attributes: %w", err)
	}

	return nil
}
