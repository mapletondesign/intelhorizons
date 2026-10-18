import satori from 'satori'
import { Resvg } from '@resvg/resvg-js'
import fs from 'fs'
import path from 'path'

const API_URL = process.env.API_URL ?? 'https://intelhorizons.fly.dev'
const OUTPUT_DIR = './output'
const WIDTH = 1080
const HEIGHT = 1080

// ── CLI args ──────────────────────────────────────────────────────────────────

const args = process.argv.slice(2)
const limitArg = args.includes('--limit') ? parseInt(args[args.indexOf('--limit') + 1]) : 5
const categoryArg = args.includes('--category') ? args[args.indexOf('--category') + 1] : undefined

// ── Types ─────────────────────────────────────────────────────────────────────

interface Article {
  ID: number
  Title: string
  Summary: string
  URL: string
  SourceName: string
  Score: number
  Category: string
  PublishedAt: string
}

// ── Category colours ──────────────────────────────────────────────────────────

const categoryStyle: Record<string, { bg: string; text: string }> = {
  'Claude':                { bg: '#312e81', text: '#a5b4fc' },
  'GPT / OpenAI':          { bg: '#064e3b', text: '#6ee7b7' },
  'Frontier Models':       { bg: '#042f2e', text: '#5eead4' },
  'Open Source Models':    { bg: '#451a03', text: '#fcd34d' },
  'Agents & Automation':   { bg: '#0c4a6e', text: '#7dd3fc' },
  'Safety & Policy':       { bg: '#4c0519', text: '#fda4af' },
  'Research & Benchmarks': { bg: '#2e1065', text: '#c4b5fd' },
  'Image & Video':         { bg: '#500724', text: '#f9a8d4' },
}

function scoreColor(score: number): string {
  if (score >= 7) return '#6366f1'
  if (score >= 4) return '#f59e0b'
  return '#64748b'
}

// ── Font loader ───────────────────────────────────────────────────────────────

async function loadGoogleFont(family: string, weight: number): Promise<ArrayBuffer> {
  // Use an old user agent so Google Fonts returns TTF (truetype) instead of woff2.
  // Satori's opentype.js does not support woff2.
  const css = await fetch(
    `https://fonts.googleapis.com/css2?family=${encodeURIComponent(family)}:wght@${weight}`,
    { headers: { 'User-Agent': 'Mozilla/5.0' } }
  ).then(r => r.text())

  const urls = [...css.matchAll(/src: url\((.+?)\) format\('truetype'\)/g)].map(m => m[1])
  const url = urls.at(-1)
  if (!url) throw new Error(`No truetype URL found for ${family}:${weight}`)
  return fetch(url).then(r => r.arrayBuffer())
}

// ── Card template ─────────────────────────────────────────────────────────────

function buildCard(article: Article) {
  const cat = article.Category
  const catStyle = categoryStyle[cat]
  const filled = Math.round((article.Score / 10) * 5)
  const sc = scoreColor(article.Score)

  const title = article.Title.length > 100
    ? article.Title.slice(0, 97) + '…'
    : article.Title

  const summary = article.Summary.length > 160
    ? article.Summary.slice(0, 157) + '…'
    : article.Summary

  return {
    type: 'div',
    props: {
      style: {
        display: 'flex',
        flexDirection: 'column',
        width: '100%',
        height: '100%',
        backgroundColor: '#0f172a',
        padding: '80px',
        fontFamily: 'Inter',
      },
      children: [

        // ── Header: brand + category ─────────────────────────────────────────
        {
          type: 'div',
          props: {
            style: { display: 'flex', justifyContent: 'space-between', alignItems: 'center' },
            children: [
              {
                type: 'div',
                props: {
                  style: { display: 'flex', alignItems: 'center', gap: '10px' },
                  children: [
                    {
                      type: 'div',
                      props: {
                        style: {
                          display: 'flex',
                          width: '8px', height: '8px',
                          borderRadius: '50%',
                          backgroundColor: '#6366f1',
                        },
                        children: [],
                      },
                    },
                    {
                      type: 'span',
                      props: {
                        style: { fontSize: '18px', fontWeight: 700, color: '#e2e8f0', letterSpacing: '0.04em', textTransform: 'uppercase' },
                        children: 'IntelHorizons',
                      },
                    },
                  ],
                },
              },
              catStyle
                ? {
                    type: 'div',
                    props: {
                      style: {
                        display: 'flex',
                        backgroundColor: catStyle.bg,
                        color: catStyle.text,
                        borderRadius: '99px',
                        padding: '6px 18px',
                        fontSize: '13px',
                        fontWeight: 600,
                        letterSpacing: '0.06em',
                        textTransform: 'uppercase',
                      },
                      children: cat,
                    },
                  }
                : { type: 'div', props: { style: { display: 'flex' }, children: [] } },
            ],
          },
        },

        // ── Divider ──────────────────────────────────────────────────────────
        {
          type: 'div',
          props: {
            style: { display: 'flex', width: '100%', height: '1px', backgroundColor: '#1e293b', marginTop: '48px' },
            children: [],
          },
        },

        // ── Title ────────────────────────────────────────────────────────────
        {
          type: 'div',
          props: {
            style: {
              display: 'flex',
              marginTop: '64px',
              fontSize: '58px',
              fontWeight: 700,
              color: '#f1f5f9',
              lineHeight: 1.15,
              letterSpacing: '-0.02em',
            },
            children: title,
          },
        },

        // ── Summary ──────────────────────────────────────────────────────────
        summary
          ? {
              type: 'div',
              props: {
                style: {
                  display: 'flex',
                  marginTop: '32px',
                  fontSize: '26px',
                  color: '#64748b',
                  lineHeight: 1.5,
                  fontWeight: 400,
                },
                children: summary,
              },
            }
          : { type: 'div', props: { style: { display: 'flex' }, children: [] } },

        // ── Spacer ───────────────────────────────────────────────────────────
        { type: 'div', props: { style: { display: 'flex', flex: 1 }, children: [] } },

        // ── Footer ───────────────────────────────────────────────────────────
        {
          type: 'div',
          props: {
            style: {
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              paddingTop: '32px',
              borderTop: '1px solid #1e293b',
            },
            children: [

              // Score bar
              {
                type: 'div',
                props: {
                  style: { display: 'flex', alignItems: 'center', gap: '12px' },
                  children: [
                    {
                      type: 'div',
                      props: {
                        style: { display: 'flex', gap: '5px' },
                        children: Array.from({ length: 5 }, (_, i) => ({
                          type: 'div',
                          props: {
                            style: {
                              display: 'flex',
                              width: '28px',
                              height: '6px',
                              borderRadius: '3px',
                              backgroundColor: i < filled ? sc : '#1e293b',
                            },
                            children: [],
                          },
                        })),
                      },
                    },
                    {
                      type: 'span',
                      props: {
                        style: { fontSize: '18px', color: '#475569', fontWeight: 600 },
                        children: article.Score.toFixed(1),
                      },
                    },
                  ],
                },
              },

              // Source + URL
              {
                type: 'div',
                props: {
                  style: { display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '4px' },
                  children: [
                    {
                      type: 'span',
                      props: {
                        style: { fontSize: '18px', color: '#94a3b8', fontWeight: 600 },
                        children: article.SourceName,
                      },
                    },
                    {
                      type: 'span',
                      props: {
                        style: { fontSize: '14px', color: '#334155' },
                        children: 'intelhorizons.com',
                      },
                    },
                  ],
                },
              },

            ],
          },
        },

      ],
    },
  }
}

// ── Main ──────────────────────────────────────────────────────────────────────

async function main() {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true })

  console.log('Loading fonts…')
  const [fontRegular, fontBold] = await Promise.all([
    loadGoogleFont('Inter', 400),
    loadGoogleFont('Inter', 700),
  ])

  const params = new URLSearchParams({ limit: String(limitArg) })
  if (categoryArg) params.set('category', categoryArg)

  console.log(`Fetching articles from ${API_URL}…`)
  const articles: Article[] = await fetch(`${API_URL}/api/feed?${params}`).then(r => r.json())

  if (!articles.length) {
    console.log('No articles returned.')
    return
  }

  console.log(`Generating ${articles.length} card(s)…`)

  for (const article of articles) {
    const svg = await satori(buildCard(article) as any, {
      width: WIDTH,
      height: HEIGHT,
      fonts: [
        { name: 'Inter', data: fontRegular, weight: 400, style: 'normal' },
        { name: 'Inter', data: fontBold,    weight: 700, style: 'normal' },
      ],
    })

    const resvg = new Resvg(svg, { fitTo: { mode: 'width', value: WIDTH } })
    const png = resvg.render().asPng()

    const slug = article.Title
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .slice(0, 50)
      .replace(/-$/, '')
    const filename = `${article.ID}-${slug}.png`
    const outPath = path.join(OUTPUT_DIR, filename)

    fs.writeFileSync(outPath, png)
    console.log(`  ✓ ${filename}  (score ${article.Score.toFixed(1)}, ${article.Category || 'uncategorised'})`)
  }

  console.log(`\nDone — cards saved to ${path.resolve(OUTPUT_DIR)}`)
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})
