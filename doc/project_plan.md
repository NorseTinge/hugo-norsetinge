# Projektdokument: Norsetinge Nyhedstjeneste

Dette dokument beskriver arkitekturen og arbejdsgangen for den automatiserede, flersprogede nyhedstjeneste Norsetinge.

## 1. Vision

At skabe en højt automatiseret nyhedsplatform, der kan modtage en enkelt artikel, oversætte den til en bred vifte af sprog, og publicere den globalt med minimal manuel indgriben. Systemet skal være robust, skalerbart og let at vedligeholde.

## 2. Overordnet Arkitektur

Systemet består af tre hovedkomponenter:

1.  **Input (Dropbox):** Et simpelt startpunkt, hvor en ny artikel i Markdown-format udløser hele processen.
2.  **Processor (LXC Container "norsetinge"):** Systemets hjerte. En Go-applikation orkestrerer **godkendelse**, **oversættelse**, og generering af det statiske website.
3.  **Output (Webhost `norsetinge.com`):** Det færdige, statiske Hugo-site, der serveres til brugerne.

```
┌───────────┐      ┌───────────────────────────────────┐      ┌──────────────────┐
│ Dropbox   │      │ LXC Container ("norsetinge")        │      │ Webhost          │
│           ├─────►│ 1. Watcher-script                   ├─────►│                  │
│ `udgiv/`  │      │ 2. Go-app (Klargøring & Godkendelse)│      │ norsetinge.com   │
│ article.md│      │ 3. API-kald (Oversæt EFTER godk.)   │      │ (Statisk site)   │
│           │      │ 4. Hugo (Site-generator)            │      │                  │
└───────────┘      └───────────────────────────────────┘      └──────────────────┘
```

## 3. Detaljeret Arbejdsgang (Pipeline)

### Trin 1: Indhold og Udløser

*   **Kilde:** En bruger placerer en Markdown-fil i `dropbox/sti/norsetinge/udgiv/`.
*   **Format:** Filen **skal** indeholde YAML frontmatter med minimum `title` og `author`. F.eks. `author: "TB (twisted brain)"`.
*   **Udløser:** Et `watch`-script på LXC-containeren "norsetinge" opdager den nye fil.

### Trin 2: Godkendelse (FØR oversættelse)

*   **Mekanisme:** En simpel web-app, der kører på LXC-containeren og er tilgængelig via et **Tailscale-netværkslink (tailnet)**.
*   **Proces:**
    1.  Go-applikationen læser den originale artikel.
    2.  Den genererer et preview af artiklen i web-appen.
    3.  Systemet sender en e-mail til dig med et direkte link til preview-siden på tailnet'et.
*   **Handlinger på Godkendelsessiden:**
    *   **Godkend:** Processen fortsætter til Trin 3 (Oversættelse).
    *   **Foreslå Ændringer:** En tekstboks lader dig skrive kommentarer. Når du sender, stoppes processen, og den originale fil flyttes til en mappe som `dropbox/sti/norsetinge/afventer-rettelser/` for manuelt gennemsyn.
    *   **Afvis:** Processen stoppes, og filen flyttes til `dropbox/sti/norsetinge/afvist/`.

### Trin 3: Oversættelse

*   **Udløser:** En succesfuld godkendelse fra Trin 2.
*   **Værktøj:** Go-applikationen.
*   **Proces:**
    1.  Go-applikationen tager den godkendte originalartikel.
    2.  Den kalder **OpenRouter API** for at oversætte artiklen til alle definerede målsprog.
*   **Sprog:**
    *   **Primært sprog:** Engelsk (`en`).
    *   **Kernesprog:** Dansk (`da`), Svensk (`se`), Norsk (`no`), Finsk (`fi`), Tysk (`de`), Fransk (`fr`), Italiensk (`it`), Spansk (`es`), Græsk (`gr`), Grønlandsk (`kl`), Islandsk (`is`), Færøsk (`fo`).
    *   **Europæiske sprog:** Russisk (`ru`), Tyrkisk (`tr`), Ukrainsk (`uk`), Estisk (`et`), Lettisk (`lv`), Litauisk (`lt`) etc.
    *   **Asiatiske sprog:** Kinesisk (`zh`), Koreansk (`ko`), Japansk (`ja`).

### Trin 4: Generering af Website

*   **Værktøj:** Hugo.
*   **Proces:**
    1.  Go-applikationen tager originalen og alle oversættelserne.
    2.  Den opdaterer `hugin.json`-kontrakten og placerer alle Markdown-filerne (`index.da.md`, `index.en.md` etc.) i den korrekte Hugo-mappestruktur.
    3.  Hugo køres for at bygge hele det statiske website til en `public/`-mappe.

### Trin 5: Publicering og Arkivering

*   **Publicering:** Indholdet af `public/`-mappen synkroniseres til webhotellet via `rsync` eller lignende.
*   **Arkivering:** Efter succesfuld publicering udfører Go-applikationen følgende:
    1.  Den færdige artikels URL bestemmes (f.eks. `https://norsetinge.com/artikel/slug-name/`).
    2.  Den originale Markdown-fil fra `udgiv/` flyttes til `dropbox/sti/norsetinge/udgivet/`.
    3.  Go-applikationen **indsætter nyt metadata** i den flyttede fil, f.eks.:
        ```yaml
        publication_date: "2025-10-02T10:00:00Z"
        publication_url: "https://norsetinge.com/artikel/slug-name/"
        ```

## 4. Specifikationer for Websitet

*   **URL-struktur:** `https://norsetinge.com/artikel/{sprogkode}/{artikel-slug}/`.
    *   Engelsk er standard og kan tilgås uden sprogkode: `https://norsetinge.com/artikel/{artikel-slug}/`.
*   **Sproghåndtering (Frontend):**
    *   En JavaScript-fil (`assets/js/language.js`) håndterer:
        1.  **Auto-detektering:** Aflæser browserens sprog (`navigator.language`).
        2.  **Redirect:** Viderestiller brugeren til den korrekte sprogversion, hvis den findes.
        3.  **Sprogvælger:** En dropdown-menu, der lader brugeren vælge sprog manuelt.
        4.  **Cookie:** Når en bruger aktivt vælger et sprog, gemmes valget i en cookie (`user_lang_preference`). Ved næste besøg prioriteres dette valg over browserens sprog.
*   **Cookie-samtykke:** I sprogvælgeren vises teksten: "Vi bruger cookies til at gemme dit sprogvalg." Denne tekst skal vises på det sprog, browseren er indstillet til.
*   **Footer:** Skal indeholde:
    *   Udgivelsesdato.
    *   Forfatter (hentes fra `author`-feltet i sidens front matter, f.eks. "TB (twisted brain)").
    *   Copyright-notits med årstal.
*   **Feedback:** En sektion i footeren linker til GitHub-projektets "Issues"-side, så brugere kan indmelde fejl og forslag.

## 5. Teknologivalg

*   **Backend / Pipeline:** **Go** (primært), shell-scripts.
*   **Oversættelse:** **OpenRouter API**.
*   **Godkendelses-app:** **Go** (net/http), **Node.js** eller **PHP** (alle er mulige på hosten).
*   **Site Generator:** **Hugo**.
*   **Frontend:** **HTML**, **CSS**, **TypeScript/JavaScript** (kun til sprogfunktionalitet).
*   **Hosting:** Standard webhost, der kan servere statiske filer.

## 6. Forslag til Forbedringer

*   **Internt Annoncesystem:**
    *   **Idé:** Vis egne annoncer/meddelelser på tværs af sitet.
    *   **Implementering:** Opret en `data/ads.yaml`-fil i Hugo. Her kan annoncer defineres med tekst, billede og link. Et "partial" i Hugo-temaet kan så hente en tilfældig annonce fra denne fil og vise den på relevante sider.
*   **Avanceret CLI-værktøj:**
    *   **Idé:** Udbyg `hugin`-værktøjet fra det oprindelige oplæg.
    *   **Funktioner:** `hugin new article --file "path/to/local.md"`, `hugin build --force`, `hugin deploy`. Dette giver en alternativ arbejdsgang udenom Dropbox.
*   **Pipeline Dashboard:**
    *   **Idé:** En simpel web-side på LXC-containeren, der viser status på artikler i pipelinen: "Modtaget", "Oversætter", "Venter på godkendelse", "Publiceret".

---
