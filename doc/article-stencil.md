# Article Stencil - Complete Guide
**Version:** 1.0
**Dato:** 2025-10-03

---

## Formål

Dette dokument beskriver alle aspekter af artikel-skabelonen og hvordan artikler struktureres i Norsetinge systemet, inklusiv alle felter understøttet af koden og avancerede Hugo shortcodes.

---

## Artikel Struktur

### Komplet Format

```markdown
---
# AUTO-GENERATED (do not edit manually)
id: "#ABC123"

# REQUIRED FIELDS
title: "Din artikel-titel her"
author: "TB (twisted brain)"
status:
  draft: 0
  revision: 0
  publish: 0      # Sæt til 1 når klar til godkendelse
  published: 0
  rejected: 0
  update: 0

# SEO & SOCIAL MEDIA (anbefalet)
description: "Kort resume af artiklen (anbefalet for SEO og social media)"
images:
  - "/images/hero-image.jpg"
  - "/images/secondary-image.jpg"

# ORGANIZATION (valgfrit)
slug: "custom-url-slug"  # Auto-genereret fra titel hvis ikke angivet
tags: ["tag1", "tag2", "tag3"]
categories: ["Kategori1", "Kategori2"]
series: "Serie Navn"  # Hvis del af en artikel-serie

# MEDIA (valgfrit)
videos:
  - "https://youtube.com/watch?v=xyz"
  - "/videos/local-video.mp4"
audio:
  - "/audio/podcast-episode.mp3"

# BRANDING (valgfrit)
favicon: "/images/custom-favicon.png"
app_icon: "/images/custom-app-icon.png"
---

# Artikel Indhold Starter Her

Dit markdown indhold kommer EFTER `---` afslutningen.

## Underoverskrift

Tekst, billeder, shortcodes, osv.
```

---

## Frontmatter Felter - Detaljeret

### 1. Auto-Generated Fields

#### `id` (auto-genereret)

**Type:** `string`
**Format:** `"#ABC123"` (6 hex chars med # prefix)
**Auto-genereret:** Ja
**Persistens:** Permanent (skrevet til fil ved første parse)

```yaml
id: "#A3F2E9"
```

**Formål:**
- Unik identifier for artikel gennem hele lifecycle
- Bruges i approval URLs
- Tracker pending approvals
- Linking mellem artikler

**Generering:**
```go
// Auto-genereret ved første load hvis mangler
func (a *Article) generateID() string {
    hash := sha256.Sum256([]byte(a.Title + a.Author + time.Now().String()))
    return "#" + hex.EncodeToString(hash[:3])
}
```

**⚠️ VIGTIGT:** Rediger ALDRIG manuelt. Systemet genererer automatisk.

---

### 2. Required Fields

#### `title` (påkrævet)

**Type:** `string`
**Max længde:** Anbefalet 100 chars
**SEO impact:** Høj

```yaml
title: "DevOps som paradigme: Samarbejde, kultur og inklusion"
```

**Anvendelse:**
- Vises som `<h1>` i artikel
- Bruges i `<title>` tag
- Vises i ntfy notifikationer
- Bruges til auto-generering af slug

**Best practices:**
- Klar og beskrivende
- Inkluder nøgleord for SEO
- Undgå specialtegn der kan give problemer i URLs
- Max 60-70 chars for optimal SEO

---

#### `author` (påkrævet)

**Type:** `string`
**Format:** Fri tekst eller initialer

```yaml
author: "TB (twisted brain)"
# Eller
author: "Lars Peter Mathiasen"
```

**Anvendelse:**
- Vises i artikel byline
- Inkluderet i ntfy notifikationer
- Copyright attribution
- SEO author tag

---

#### `status` (påkrævet)

**Type:** `object` med 6 integer flags
**System-kritisk:** Ja - bestemmer workflow

```yaml
status:
  draft: 0      # Work in progress
  revision: 0   # Needs revision
  publish: 0    # Ready for approval
  published: 0  # Approved and live
  rejected: 0   # Rejected by editor
  update: 0     # Update to existing article
```

**Regler:**
- **Alle `0`:** Systemet ignorerer artiklen (safe draft mode)
- **Mindst én `1`:** Systemet reagerer
- **"Last 1 wins":** Hvis flere er `1`, tæller den sidste i rækkefølgen

**Workflow:**

```
kladde/         draft: 0, publish: 0     → Systemet ignorerer
                     ↓
                Sæt publish: 1
                     ↓
udgiv/          publish: 1               → Preview + Ntfy
                     ↓
             Editor godkender
                     ↓
udgivet/        published: 1             → Inkluderet i build
                     ↓
                  Deploy
                     ↓
              Live på webhost
```

**Status Transitions:**

| Fra Folder | Status Before | Action | Status After | Til Folder |
|------------|---------------|--------|--------------|------------|
| kladde/    | all 0         | Ready  | publish: 1   | udgiv/     |
| udgiv/     | publish: 1    | Approve| published: 1 | udgivet/   |
| udgiv/     | publish: 1    | Reject | rejected: 1  | afvist/    |
| afvist/    | rejected: 1   | Fix    | update: 1    | udgiv/     |

---

### 3. SEO & Social Media Fields (anbefalet)

#### `description` (anbefalet)

**Type:** `string`
**Max længde:** 155-160 chars (SEO optimal)
**SEO impact:** Høj

```yaml
description: "DevOps er mere end teknologi – det er et paradigme, hvor teams arbejder sammen på tværs af siloer med fokus på kultur, samarbejde og inklusion."
```

**Anvendelse:**
- Meta description tag (`<meta name="description">`)
- Open Graph description (`og:description`)
- Twitter Card description
- Preview i søgemaskiner
- Lead paragraph i artikel

**Best practices:**
- Skriv actionable og engaging
- Inkluder primært nøgleord
- Max 155 chars for at undgå trunkering i Google
- Undgå duplikering af title

---

#### `images` (anbefalet)

**Type:** `array` af strings (URLs/paths)
**SEO impact:** Mellem
**Social media impact:** Høj

```yaml
images:
  - "/images/devops-team-collaboration.jpg"
  - "/images/devops-workflow-diagram.png"
```

**Anvendelse:**
- Open Graph image (`og:image`)
- Twitter Card image
- JSON-LD structured data
- Featured image i article header
- Gallery/lightbox hvis flere

**Best practices:**
- Første image bruges som primary featured image
- Min størrelse: 1200x630px (Open Graph optimal)
- Format: JPG, PNG, eller WebP
- Alt text tilføjes via `{{< figure >}}` shortcode

---

### 4. Organization Fields (valgfrit)

#### `slug` (valgfrit)

**Type:** `string`
**Auto-genereret:** Ja (fra title hvis ikke angivet)
**URL impact:** Direkte

```yaml
slug: "devops-paradigme-samarbejde"
```

**Anvendelse:**
- URL path: `https://norsetinge.com/articles/devops-paradigme-samarbejde/`
- Permanent link til artikel

**Auto-generation:**
```go
// Hvis slug ikke angivet, genereres fra title:
// "DevOps som paradigme" → "devops-som-paradigme"
slug := strings.ToLower(title)
slug = strings.ReplaceAll(slug, " ", "-")
slug = removeSpecialChars(slug)
```

**Best practices:**
- Kort og beskrivende
- Kun lowercase, tal, og bindestreger
- Undgå danske tegn (æ, ø, å) → brug (ae, oe, aa)
- Max 50-60 chars

---

#### `tags` (valgfrit)

**Type:** `array` af strings
**SEO impact:** Lav-mellem
**Organization impact:** Høj

```yaml
tags:
  - "DevOps"
  - "Paradigme"
  - "Teamarbejde"
  - "Inklusion"
  - "Kultur"
```

**Anvendelse:**
- Tag archive pages (`/tags/devops/`)
- Related articles matching
- Search/filter functionality
- SEO keywords (lav vægt)

**Best practices:**
- 3-7 tags per artikel
- Konsistent naming (check existing tags)
- Brug CamelCase eller lowercase
- Specifik > generel

---

#### `categories` (valgfrit)

**Type:** `array` af strings
**SEO impact:** Mellem
**Organization impact:** Høj

```yaml
categories:
  - "Teknologi"
  - "Ledelse"
  - "Arbejdskultur"
```

**Anvendelse:**
- Category archive pages (`/categories/teknologi/`)
- Main navigation
- Article grouping
- RSS feeds per category

**Best practices:**
- 1-3 categories per artikel
- Brug eksisterende categories (check først)
- Hierarkisk struktur mulig

---

#### `series` (valgfrit)

**Type:** `string`
**Organization impact:** Høj

```yaml
series: "DevOps i Praksis"
```

**Anvendelse:**
- Link relaterede artikler i samme serie
- "Next/Previous in series" navigation
- Series overview page

**Eksempel serie:**
```yaml
# Article 1
title: "DevOps: Introduktion"
series: "DevOps Guide"

# Article 2
title: "DevOps: Værktøjer"
series: "DevOps Guide"

# Article 3
title: "DevOps: Best Practices"
series: "DevOps Guide"
```

---

### 5. Media Fields (valgfrit)

#### `videos` (valgfrit)

**Type:** `array` af strings (URLs)
**Formats:** YouTube, Vimeo, local MP4

```yaml
videos:
  - "https://youtube.com/watch?v=dQw4w9WgXcQ"
  - "https://vimeo.com/123456789"
  - "/videos/interview-2025.mp4"
```

**Anvendelse:**
- Auto-embed via shortcodes
- Video sitemap for SEO
- Schema.org VideoObject

**Hugo rendering:**
```markdown
{{< youtube dQw4w9WgXcQ >}}
{{< vimeo 123456789 >}}
{{< video src="/videos/interview.mp4" >}}
```

---

#### `audio` (valgfrit)

**Type:** `array` af strings (paths/URLs)
**Formats:** MP3, OGG, WAV

```yaml
audio:
  - "/audio/podcast-episode-1.mp3"
  - "/audio/interview.mp3"
```

**Anvendelse:**
- Embedded audio player
- Podcast RSS feed
- Schema.org AudioObject

**Hugo rendering:**
```markdown
{{< audio src="/audio/podcast-episode-1.mp3" >}}
```

---

### 6. Branding Fields (valgfrit)

#### `favicon` (valgfrit)

**Type:** `string` (path)
**Format:** ICO, PNG (16x16, 32x32, 48x48)

```yaml
favicon: "/images/custom-favicon.png"
```

**Anvendelse:**
- Article-specific favicon
- Override site default

---

#### `app_icon` (valgfrit)

**Type:** `string` (path)
**Format:** PNG (180x180, 192x192, 512x512)

```yaml
app_icon: "/images/custom-app-icon.png"
```

**Anvendelse:**
- Apple Touch Icon
- Android homescreen icon
- PWA manifest icon

---

## Markdown Content (Efter Frontmatter)

### Struktur

```markdown
---
frontmatter: here
---

# H1 Hovedoverskrift (optional - title bruges automatisk)

Lead paragraph - første afsnit der opsummerer artiklen.

## H2 Første sektion

Tekst indhold med **fed**, *kursiv*, og [links](https://example.com).

### H3 Undersektion

- Bullet point liste
- Punkt to
- Punkt tre

1. Nummereret liste
2. Punkt to
3. Punkt tre

> Blockquote for citat

`Inline code`

```
Kodeblok
```

![Alt text](/images/billede.jpg)
```

---

## Hugo Shortcodes

### Indbyggede (virker med det samme)

#### YouTube Video

```markdown
{{< youtube dQw4w9WgXcQ >}}
```

**Parameters:**
- Video ID (påkrævet)
- `autoplay=1` (valgfrit)
- `start=30` (valgfrit - start ved 30 sekunder)

---

#### Vimeo Video

```markdown
{{< vimeo 123456789 >}}
```

---

#### Figure (Billede med Caption)

```markdown
{{< figure
  src="/images/billede.jpg"
  alt="Billedbeskrivelse for screen readers"
  caption="Dette er en billedtekst"
  width="800"
  link="https://example.com"
>}}
```

**Parameters:**
- `src` - Image path (påkrævet)
- `alt` - Alt text for accessibility (påkrævet)
- `caption` - Billedtekst (valgfrit)
- `title` - Hover tooltip (valgfrit)
- `width` / `height` - Dimensioner (valgfrit)
- `link` - URL ved klik (valgfrit)
- `class` - CSS class (valgfrit)

---

#### Twitter Embed

```markdown
{{< twitter user="username" id="1234567890" >}}
```

---

#### GitHub Gist

```markdown
{{< gist brugernavn gist_id >}}
```

---

### Custom Shortcodes (Norsetinge-specifikke)

Disse implementeres i `site/layouts/shortcodes/` når Hugo theme bygges.

#### Gallery (Billedgalleri)

```markdown
{{< gallery >}}
  {{< img src="/images/billede1.jpg" alt="Beskrivelse 1" >}}
  {{< img src="/images/billede2.jpg" alt="Beskrivelse 2" >}}
  {{< img src="/images/billede3.jpg" alt="Beskrivelse 3" >}}
{{< /gallery >}}
```

**Features:**
- Lightbox/modal view
- Swipe navigation på mobil
- Lazy loading

---

#### Highlight Box

```markdown
{{< highlight-box type="info" >}}
Dette er vigtig information som fremhæves i artiklen.
{{< /highlight-box >}}
```

**Types:**
- `info` - Blå (ℹ️ information)
- `warning` - Gul (⚠️ advarsel)
- `success` - Grøn (✅ success)
- `error` - Rød (❌ fejl)

**Rendering:**
```html
<div class="highlight-box highlight-box--info">
  <div class="highlight-box__icon">ℹ️</div>
  <div class="highlight-box__content">
    Dette er vigtig information...
  </div>
</div>
```

---

#### Pull Quote

```markdown
{{< pullquote author="Navn Navnesen" title="CEO, Firma" >}}
Dette er et vigtigt citat fra artiklen som skal fremhæves.
{{< /pullquote >}}
```

**Features:**
- Large, styled quote
- Author attribution
- Optional title/position

---

#### Audio Player

```markdown
{{< audio src="/audio/interview.mp3" title="Interview med Expert" >}}
```

---

#### Video Player (Local)

```markdown
{{< video
  src="/videos/news-clip.mp4"
  poster="/images/video-thumbnail.jpg"
  width="800"
>}}
```

**Features:**
- HTML5 video player
- Custom poster frame
- Controls

---

#### Related Articles

```markdown
{{< related-articles >}}
artikel-slug-1
artikel-slug-2
artikel-slug-3
{{< /related-articles >}}
```

**Auto-generates:**
- Links to specified articles
- Thumbnails if available
- Excerpt previews

---

#### Info Box

```markdown
{{< infobox icon="lightbulb" title="Husk dette" >}}
Vigtigt information læseren skal være opmærksom på.
{{< /infobox >}}
```

**Icons:** lightbulb, warning, info, check, cross

---

#### Accordion (Sammenfoldbar)

```markdown
{{< accordion title="Klik for at læse mere" open="false" >}}
Dette indhold er skjult indtil brugeren klikker.

Kan indeholde **markdown** og andet content.
{{< /accordion >}}
```

**Parameters:**
- `title` - Clickable header (påkrævet)
- `open` - Start expanded (default: false)

---

#### Sortable Table

```markdown
{{< table sortable="true" class="striped" >}}
| Navn | Alder | By |
|------|-------|-----|
| Anna | 32 | København |
| Peter | 45 | Aarhus |
| Marie | 28 | Odense |
{{< /table >}}
```

**Features:**
- Click column headers to sort
- Striped rows
- Responsive (scroll on mobile)

---

## Komplet Artikel Eksempel

```markdown
---
title: "DevOps som paradigme: Samarbejde, kultur og inklusion"
author: "TB (twisted brain)"
description: "DevOps er mere end teknologi – det er et paradigme, hvor teams arbejder sammen på tværs af siloer."
slug: "devops-paradigme"
tags:
  - "DevOps"
  - "Paradigme"
  - "Teamarbejde"
  - "Kultur"
categories:
  - "Teknologi"
  - "Ledelse"
series: "DevOps i Praksis"
images:
  - "/images/devops-collaboration.jpg"
status:
  draft: 0
  revision: 0
  publish: 1  # KLAR TIL GODKENDELSE
  published: 0
  rejected: 0
  update: 0
---

Når vi taler om DevOps, reduceres det ofte til CI/CD-pipelines, Kubernetes eller automatisering. Men DevOps er først og fremmest et **paradigme** – en måde at tænke og arbejde på.

{{< figure
  src="/images/devops-collaboration.jpg"
  alt="DevOps team collaboration"
  caption="DevOps teams arbejder på tværs af siloer"
>}}

{{< highlight-box type="info" >}}
**DevOps Definition:** Development + Operations = Kultur af samarbejde
{{< /highlight-box >}}

## Fra siloer til fælles ansvar

I mange organisationer er udviklere og driftsfolk adskilt – med modsatrettede mål.

{{< pullquote author="Gene Kim" title="Forfatter, The Phoenix Project" >}}
DevOps er fundamentalt omkring at bryde ned de mure der eksisterer mellem development og operations.
{{< /pullquote >}}

### Fælles mål

* **Deployment frequency** - Hvor ofte shipper vi?
* **Lead time** - Hvor hurtigt fra idé til produktion?
* **MTTR** - Mean time to recovery
* **Change failure rate** - Hvor mange deployments fejler?

{{< table sortable="true" >}}
| Metric | Traditional | DevOps |
|--------|-------------|--------|
| Deploy frequency | Monthly | Daily/Hourly |
| Lead time | Weeks | Hours |
| MTTR | Hours | Minutes |
| Change failure | 30% | <5% |
{{< /table >}}

## Inklusion i praksis

{{< accordion title="Hvad betyder inklusion i DevOps?" >}}
Inklusion handler om at inkludere ALLE roller i processen:

- Udvikling
- Drift
- Test
- Sikkerhed (DevSecOps)
- Design
- Product
- Slutbrugere
{{< /accordion >}}

## Video: DevOps Explained

{{< youtube Xrgk023l4lI >}}

## Se også

{{< related-articles >}}
devops-tools-2025
continuous-deployment-guide
infrastructure-as-code
{{< /related-articles >}}

---

**Læs mere:**
- [The Phoenix Project](https://example.com)
- [DevOps Handbook](https://example.com)
```

---

## Best Practices

### Frontmatter

✅ **DO:**
- Inkluder ALTID `title`, `author`, `status`
- Tilføj `description` for SEO (155 chars)
- Brug `images` array for social sharing
- Konsistent tag/category naming
- Skriv descriptive slugs

❌ **DON'T:**
- Rediger `id` feltet manuelt
- Glem status flags (alle 0 = systemet ignorerer)
- Brug for mange tags (>10)
- Duplikér content i description og lead paragraph
- Brug specialtegn i slugs

---

### Content

✅ **DO:**
- Start med strong lead paragraph
- Brug H2/H3 struktur (ikke H1 - title er H1)
- Inkluder billeder med alt text
- Brug shortcodes for rich media
- Skriv mobile-friendly (korte paragraffer)

❌ **DON'T:**
- Embed råt HTML (brug shortcodes)
- Lange paragraffer (>5 linjer)
- Manglende alt text på billeder
- For mange embeds (slow page load)

---

### Workflow

1. **Kladde:** Skriv i `kladde/` med `status: draft: 0, publish: 0`
2. **Review:** Læs igennem, ret, tilføj media
3. **Ready:** Sæt `publish: 1`
4. **System:** Auto-flytter til `udgiv/`, sender ntfy
5. **Preview:** Åbn preview link fra ntfy
6. **Approve:** Klik "Godkend" eller "Godkend + Deploy Nu"
7. **Live:** Artikel bygges og deployes til webhost

---

## Validering

### Pre-publish Checklist

- [ ] `title` udfyldt og SEO-optimeret
- [ ] `author` korrekt
- [ ] `description` 155 chars, engaging
- [ ] Mindst 1 `image` (1200x630px+)
- [ ] 3-7 `tags` valgt
- [ ] 1-3 `categories` valgt
- [ ] `status: publish: 1` sat
- [ ] Lead paragraph strong (første 2-3 linjer)
- [ ] Alle billeder har alt text
- [ ] Links tjekket (ingen 404s)
- [ ] Shortcodes korrekt formateret
- [ ] Læst korrektur

---

## Code Reference

**Parser:** `/home/ubuntu/hugo-norsetinge/src/common/frontmatter.go`

**Supported fields:**
```go
type Article struct {
    FilePath string `yaml:"-"` // Internal only

    // Required
    ID     string `yaml:"id"`
    Title  string `yaml:"title"`
    Author string `yaml:"author"`
    Status Status `yaml:"status"`

    // SEO
    Description string   `yaml:"description,omitempty"`
    Images      []string `yaml:"images,omitempty"`

    // Organization
    Slug       string   `yaml:"slug,omitempty"`
    Tags       []string `yaml:"tags,omitempty"`
    Categories []string `yaml:"categories,omitempty"`
    Series     string   `yaml:"series,omitempty"`

    // Media
    Videos []string `yaml:"videos,omitempty"`
    Audio  []string `yaml:"audio,omitempty"`

    // Branding
    Favicon string `yaml:"favicon,omitempty"`
    AppIcon string `yaml:"app_icon,omitempty"`

    // Content
    Content string
}
```

---

*Dette dokument definerer den komplette artikel struktur og alle muligheder i Norsetinge systemet.*
