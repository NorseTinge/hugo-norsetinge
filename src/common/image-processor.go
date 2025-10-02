package common

// Image processor for automatic scaling and format conversion
//
// Responsibilities:
// 1. Validate input image meets minimum size requirements
// 2. Generate responsive image sizes:
//    - Thumbnail: 300x300
//    - Medium: 800x600
//    - Large: 1920x1080
//    - Open Graph: 1200x630 (social media)
// 3. Convert to multiple formats:
//    - WebP (modern browsers, best compression)
//    - JPEG (fallback)
//    - PNG (if transparency needed)
// 4. Favicon/icon generation:
//    - .ico format (16x16, 32x32, 48x48 multi-size)
//    - PNG fallbacks (for modern browsers)
//    - Apple touch icons (180x180, 152x152, 120x120, 76x76)
//    - Android icons (192x192, 512x512)
// 5. Optimize file sizes
// 6. Update article frontmatter with generated image paths

// TODO: Implement using:
// - github.com/disintegration/imaging (resize)
// - github.com/kolesa-team/go-webp (WebP conversion)
// - github.com/biessek/golang-ico (ICO format creation)
