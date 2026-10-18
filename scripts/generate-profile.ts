import satori from 'satori'
import { Resvg } from '@resvg/resvg-js'
import fs from 'fs'
import path from 'path'

const OUTPUT_DIR = './output'
const SIZE = 1080

async function loadGoogleFont(family: string, weight: number): Promise<ArrayBuffer> {
  const css = await fetch(
    `https://fonts.googleapis.com/css2?family=${encodeURIComponent(family)}:wght@${weight}`,
    { headers: { 'User-Agent': 'Mozilla/5.0' } }
  ).then(r => r.text())

  const urls = [...css.matchAll(/src: url\((.+?)\) format\('truetype'\)/g)].map(m => m[1])
  const url = urls.at(-1)
  if (!url) throw new Error(`No truetype URL found for ${family}:${weight}`)
  return fetch(url).then(r => r.arrayBuffer())
}

function buildProfile() {
  return {
    type: 'div',
    props: {
      style: {
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        width: '100%',
        height: '100%',
        backgroundColor: '#0f172a',
        fontFamily: 'Inter',
      },
      children: [

        // Indigo dot
        {
          type: 'div',
          props: {
            style: {
              display: 'flex',
              width: '120px',
              height: '120px',
              borderRadius: '50%',
              backgroundColor: '#6366f1',
              marginBottom: '52px',
            },
            children: [],
          },
        },

        // Wordmark
        {
          type: 'div',
          props: {
            style: {
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              gap: '0px',
            },
            children: [
              {
                type: 'span',
                props: {
                  style: {
                    fontSize: '128px',
                    fontWeight: 700,
                    color: '#f1f5f9',
                    letterSpacing: '-0.03em',
                    lineHeight: 1.1,
                  },
                  children: 'Intel',
                },
              },
              {
                type: 'span',
                props: {
                  style: {
                    fontSize: '128px',
                    fontWeight: 700,
                    color: '#6366f1',
                    letterSpacing: '-0.03em',
                    lineHeight: 1.1,
                  },
                  children: 'Horizons',
                },
              },
            ],
          },
        },

        // Tagline
        {
          type: 'span',
          props: {
            style: {
              fontSize: '36px',
              fontWeight: 400,
              color: '#475569',
              letterSpacing: '0.08em',
              textTransform: 'uppercase',
              marginTop: '24px',
            },
            children: 'AI News · Ranked Daily',
          },
        },

      ],
    },
  }
}

async function main() {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true })

  console.log('Loading fonts…')
  const [fontRegular, fontBold] = await Promise.all([
    loadGoogleFont('Inter', 400),
    loadGoogleFont('Inter', 700),
  ])

  console.log('Rendering profile image…')
  const svg = await satori(buildProfile() as any, {
    width: SIZE,
    height: SIZE,
    fonts: [
      { name: 'Inter', data: fontRegular, weight: 400, style: 'normal' },
      { name: 'Inter', data: fontBold,    weight: 700, style: 'normal' },
    ],
  })

  const resvg = new Resvg(svg, { fitTo: { mode: 'width', value: SIZE } })
  const png = resvg.render().asPng()

  const outPath = path.join(OUTPUT_DIR, 'profile.png')
  fs.writeFileSync(outPath, png)
  console.log(`Done — saved to ${path.resolve(outPath)}`)
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})
