// Package logback contains the logback XML renderer target implementation.
package logback

import (
	"bytes"
	"encoding/xml"

	"github.com/denglertai/outwatch/internal/config"
)

// Renderer renders normalized output config into minimal logback XML.
type Renderer struct{}

// Name returns the config target identifier for this renderer.
func (Renderer) Name() string {
	return "logback"
}

// Render generates deterministic logback XML for logger level definitions.
func (Renderer) Render(cfg config.OutputConfig) ([]byte, error) {
	// loggerXML maps one logger entry into logback XML.
	type loggerXML struct {
		XMLName xml.Name `xml:"logger"`
		Name    string   `xml:"name,attr"`
		Level   string   `xml:"level,attr"`
	}

	// configurationXML is the root logback XML document model.
	type configurationXML struct {
		XMLName xml.Name    `xml:"configuration"`
		Loggers []loggerXML `xml:"logger"`
	}

	names := config.SortedLoggerNames(cfg.Loggers)
	loggers := make([]loggerXML, 0, len(names))
	for _, name := range names {
		loggers = append(loggers, loggerXML{Name: name, Level: cfg.Loggers[name]})
	}

	payload, err := xml.MarshalIndent(configurationXML{Loggers: loggers}, "", "  ")
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBufferString(xml.Header)
	buf.Write(payload)
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}
