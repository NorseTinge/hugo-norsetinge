# Norsetinge - Automated Multilingual News Service

Automated pipeline for translating and publishing news articles in 22 languages.

## Quick Start

1. **Configure the system:**
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your credentials
   ```

2. **Build the Go application:**
   ```bash
   cd src
   go build -o norsetinge
   ```

3. **Run the pipeline:**
   ```bash
   ./src/norsetinge
   ```

4. **Publish an article:**
   - Place a Markdown file in `~/Dropbox/norsetinge/udgiv/`
   - Check email for approval link
   - Approve, and the article will be translated and published

## Documentation

- [GEMINI.md](GEMINI.md) - Project overview (English)
- [CLAUDE.md](CLAUDE.md) - Initialization and architecture
- [doc/project_plan.md](doc/project_plan.md) - Detailed architecture (Danish)

## Project Structure

```
hugo-norsetinge/
├── src/              # Go application
├── site/             # Hugo static site
├── doc/              # Documentation
└── Dropbox/          # Dropbox input/output directories
```

## Supported Languages

English, Danish, Swedish, Norwegian, Finnish, German, French, Italian, Spanish, Greek, Greenlandic, Icelandic, Faroese, Russian, Turkish, Ukrainian, Estonian, Latvian, Lithuanian, Chinese, Korean, Japanese

## License

TBD
