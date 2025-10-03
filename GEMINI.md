# GEMINI.md: Norsetinge Nyhedstjeneste

## Project Overview

This project, "Norsetinge," is an automated, multilingual news service. Its primary goal is to take a single article written in Markdown, automatically translate it into multiple languages, and publish it as a static website.

The architecture is composed of three main parts:

1.  **Input:** A designated directory (e.g., Dropbox) where new Markdown articles are placed to trigger the publication process.
2.  **Processor:** A Go application running in a containerized environment (LXC). This application orchestrates the entire workflow, including an approval process, translation via the OpenRouter API, and generation of the static website using Hugo.
3.  **Output:** The final, static website is deployed to a web host and served to users.

The main technologies used in this project are:

*   **Backend/Pipeline:** Go, Shell scripts
*   **Site Generator:** Hugo
*   **Translation:** OpenRouter API
*   **Frontend:** HTML, CSS, JavaScript (for language selection)

## Building and Running

### Go Application

To build the Go application (`norsetinge`):

```bash
cd /home/ubuntu/hugo-norsetinge
go build -o norsetinge ./src/main.go
```

To run the Go application with default config paths:

```bash
./norsetinge
```

To run with custom config paths:

```bash
./norsetinge --config /path/to/config.yaml --aliases /path/to/folder-aliases.yaml
```

### Hugo

To build the static site with Hugo:

```bash
hugo
```

To run the Hugo development server:

```bash
hugo server
```

## Development Conventions

The development workflow is centered around a GitOps-style pipeline triggered by placing a file in a specific directory.

### Article Format

*   Articles must be in Markdown format.
*   Each article file must contain YAML frontmatter with at least a `title` and `author`.

### Publication Pipeline

1.  A new Markdown article is placed in the `udgiv/` directory.
2.  A watch script detects the new file and triggers the Go application.
3.  The Go application initiates an approval process, sending an email with a link to a preview of the article.
4.  Upon approval, the article is translated into all configured languages.
5.  The Go application creates the necessary directory structure for Hugo and places the translated articles.
6.  Hugo is executed to build the static website.
7.  The generated website is published to the web host.
8.  The original article is archived with updated metadata, including the publication date and URL.
