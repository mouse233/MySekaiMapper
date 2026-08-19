import { mkdir, readFile, writeFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const docsRoot = join(root, 'docs')
const repo = 'https://github.com/mouse233/MySekaiMapper'

const routes = [
  'guide/introduction',
  'guide/quickstart',
  'guide/running-service',
  'guide/upload-api',
  'guide/report-server',
  'guide/notifications',
  'guide/cli',
  'reference/structure',
  'testing',
  'refactor-go',
  'disclaimer',
  'license'
]

const sources = [
  { source: 'README.md', locale: '', expectedHeadings: ['How it works', 'Quick start', 'Running the service', 'Upload API', 'Reqable Report Server', 'Notifications and static files', 'Command-line reference', 'Directory structure', 'Testing', 'Go refactor', 'Disclaimer', 'License'] },
  { source: 'doc/README.zh-CN.md', locale: 'zh-CN', expectedHeadings: ['工作流程', '快速上手', '运行服务', '上传接口', 'Reqable 上报服务器', '通知与静态文件', '命令行参考', '目录结构', '测试', 'Go 重构说明', '免责声明', '许可证'] },
  { source: 'doc/README.zh-TW.md', locale: 'zh-TW', expectedHeadings: ['運作方式', '快速開始', '執行服務', '上傳 API', 'Reqable Report Server', '通知與靜態檔案', '命令列參考', '目錄結構', '測試', 'Go 重構', '免責聲明', '授權條款'] },
  { source: 'doc/README.ja-JP.md', locale: 'ja-JP', expectedHeadings: ['仕組み', 'クイックスタート', 'サービスの実行', 'Upload API', 'Reqable Report Server', '通知と静的ファイル', 'コマンドラインリファレンス', 'ディレクトリ構成', 'テスト', 'Go リファクタリング', '免責事項', 'ライセンス'] },
  { source: 'doc/README.ko-KR.md', locale: 'ko-KR', expectedHeadings: ['작동 방식', '빠른 시작', '서비스 실행', '업로드 API', 'Reqable Report Server', '알림 및 정적 파일', '명령줄 참조', '디렉터리 구조', '테스트', 'Go 리팩터링', '면책 조항', '라이선스'] }
]

function splitSections(markdown, source) {
  const matches = [...markdown.matchAll(/^## (.+)$/gm)]
  if (matches.length !== routes.length) {
    throw new Error(`${source} must contain exactly ${routes.length} level-two sections; found ${matches.length}`)
  }
  if (source.expectedHeadings) {
    for (const [index, heading] of source.expectedHeadings.entries()) {
      if (matches[index][1] !== heading) throw new Error(`${source.source} section ${index + 1} is ${matches[index][1]}, expected ${heading}`)
    }
  }

  const preamble = markdown.slice(0, matches[0].index).trimEnd()
  const sections = matches.map((match, index) => {
    const end = matches[index + 1]?.index ?? markdown.length
    return markdown.slice(match.index, end).trim()
  })
  return { preamble, sections }
}

function rewriteLinks(markdown, destination, locale) {
  const depth = destination === 'index.md' ? 0 : destination.split('/').length - 1
  const relativeRoot = `${'../'.repeat(depth + (locale ? 1 : 0)) || './'}`
  let result = markdown
  for (const target of ['zh-CN', 'zh-TW', 'ja-JP', 'ko-KR']) {
    const targetPath = locale
      ? `${'../'.repeat(depth + 1)}${target}/`
      : `${'../'.repeat(depth) || './'}${target}/`
    result = result
      .replaceAll(`doc/README.${target}.md`, targetPath)
      .replaceAll(`README.${target}.md`, targetPath)
      .replaceAll(`docs/${target}/index.md`, targetPath)
  }
  result = result
    .replaceAll('(docs/index.md)', `(${relativeRoot})`)
    .replaceAll('(../README.md)', `(${relativeRoot})`)
    .replaceAll('(README.md)', `(${relativeRoot})`)
    .replaceAll('[MIT](LICENSE)', `[MIT](${repo}/blob/feat/go-rewrite/LICENSE)`)
  return result
}

async function writeGenerated(path, content, source) {
  await mkdir(dirname(path), { recursive: true })
  await writeFile(path, `<!-- GENERATED from ${source}; do not edit directly. -->\n\n${content.trim()}\n`, 'utf8')
}

for (const source of sources) {
  const markdown = await readFile(join(root, source.source), 'utf8')
  const { preamble, sections } = splitSections(markdown, source)
  const destinationRoot = source.locale ? join(docsRoot, source.locale) : docsRoot
  const prefix = source.locale ? `${source.locale}/` : ''
  await writeGenerated(join(destinationRoot, 'index.md'), rewriteLinks(preamble, 'index.md', source.locale), source.source)
  for (const [index, route] of routes.entries()) {
    const page = sections[index].replace(/^## /, '# ')
    const destination = `${route}.md`
    await writeGenerated(join(destinationRoot, destination), rewriteLinks(page, destination, source.locale), source.source)
  }
}
