import { defineConfig } from 'vitepress'

const repo = 'https://github.com/mouse233/MySekaiMapper'

const enSidebar = [
  {
    text: 'Guide',
    items: [
      { text: 'Introduction', link: '/guide/introduction' },
      { text: 'Quick Start', link: '/guide/quickstart' },
      { text: 'Running the Service', link: '/guide/running-service' },
      { text: 'Upload API', link: '/guide/upload-api' },
      { text: 'Reqable Report Server', link: '/guide/report-server' },
      { text: 'Notifications and Static Files', link: '/guide/notifications' },
      { text: 'CLI Reference', link: '/guide/cli' }
    ]
  },
  {
    text: 'Project',
    items: [
      { text: 'Directory Structure', link: '/reference/structure' },
      { text: 'Testing', link: '/testing' },
      { text: 'Go Refactor', link: '/refactor-go' },
      { text: 'Disclaimer', link: '/disclaimer' },
      { text: 'License', link: '/license' }
    ]
  }
]

const zhSidebar = [
  {
    text: '指南',
    items: [
      { text: '项目介绍', link: '/zh-CN/guide/introduction' },
      { text: '快速上手', link: '/zh-CN/guide/quickstart' },
      { text: '运行服务', link: '/zh-CN/guide/running-service' },
      { text: '上传接口', link: '/zh-CN/guide/upload-api' },
      { text: 'Reqable 上报服务器', link: '/zh-CN/guide/report-server' },
      { text: '通知与静态文件', link: '/zh-CN/guide/notifications' },
      { text: '命令行参考', link: '/zh-CN/guide/cli' }
    ]
  },
  {
    text: '项目',
    items: [
      { text: '目录结构', link: '/zh-CN/reference/structure' },
      { text: '测试', link: '/zh-CN/testing' },
      { text: 'Go 重构', link: '/zh-CN/refactor-go' },
      { text: '免责声明', link: '/zh-CN/disclaimer' },
      { text: '许可证', link: '/zh-CN/license' }
    ]
  }
]

const zhTWSidebar = [
  {
    text: '指南',
    items: [
      { text: '專案介紹', link: '/zh-TW/guide/introduction' },
      { text: '快速上手', link: '/zh-TW/guide/quickstart' },
      { text: '執行服務', link: '/zh-TW/guide/running-service' },
      { text: '上傳 API', link: '/zh-TW/guide/upload-api' },
      { text: 'Reqable 回報伺服器', link: '/zh-TW/guide/report-server' },
      { text: '通知與靜態檔案', link: '/zh-TW/guide/notifications' },
      { text: '命令列參考', link: '/zh-TW/guide/cli' }
    ]
  },
  {
    text: '專案',
    items: [
      { text: '目錄結構', link: '/zh-TW/reference/structure' },
      { text: '測試', link: '/zh-TW/testing' },
      { text: 'Go 重構', link: '/zh-TW/refactor-go' },
      { text: '免責聲明', link: '/zh-TW/disclaimer' },
      { text: '授權條款', link: '/zh-TW/license' }
    ]
  }
]

const jaSidebar = [
  {
    text: 'ガイド',
    items: [
      { text: '概要', link: '/ja-JP/guide/introduction' },
      { text: 'クイックスタート', link: '/ja-JP/guide/quickstart' },
      { text: 'サービスの実行', link: '/ja-JP/guide/running-service' },
      { text: 'アップロード API', link: '/ja-JP/guide/upload-api' },
      { text: 'Reqable レポートサーバー', link: '/ja-JP/guide/report-server' },
      { text: '通知と静的ファイル', link: '/ja-JP/guide/notifications' },
      { text: 'CLI リファレンス', link: '/ja-JP/guide/cli' }
    ]
  },
  {
    text: 'プロジェクト',
    items: [
      { text: 'ディレクトリ構成', link: '/ja-JP/reference/structure' },
      { text: 'テスト', link: '/ja-JP/testing' },
      { text: 'Go リファクタリング', link: '/ja-JP/refactor-go' },
      { text: '免責事項', link: '/ja-JP/disclaimer' },
      { text: 'ライセンス', link: '/ja-JP/license' }
    ]
  }
]

const koSidebar = [
  {
    text: '가이드',
    items: [
      { text: '소개', link: '/ko-KR/guide/introduction' },
      { text: '빠른 시작', link: '/ko-KR/guide/quickstart' },
      { text: '서비스 실행', link: '/ko-KR/guide/running-service' },
      { text: '업로드 API', link: '/ko-KR/guide/upload-api' },
      { text: 'Reqable 보고 서버', link: '/ko-KR/guide/report-server' },
      { text: '알림 및 정적 파일', link: '/ko-KR/guide/notifications' },
      { text: 'CLI 참조', link: '/ko-KR/guide/cli' }
    ]
  },
  {
    text: '프로젝트',
    items: [
      { text: '디렉터리 구조', link: '/ko-KR/reference/structure' },
      { text: '테스트', link: '/ko-KR/testing' },
      { text: 'Go 리팩터링', link: '/ko-KR/refactor-go' },
      { text: '면책 조항', link: '/ko-KR/disclaimer' },
      { text: '라이선스', link: '/ko-KR/license' }
    ]
  }
]

function localeTheme(guideText, guideLink, sidebar, editText) {
  return {
    nav: [{ text: guideText, link: guideLink }, { text: 'GitHub', link: repo }],
    sidebar,
    socialLinks: [{ icon: 'github', link: repo }],
    editLink: { pattern: `${repo}/edit/feat/go-rewrite/docs/:path`, text: editText }
  }
}

export default defineConfig({
  title: 'MySekaiMapper',
  description: 'A Go resource-gathering point map generator for MySekai',
  cleanUrls: true,
  themeConfig: localeTheme('Guide', '/guide/introduction', enSidebar, 'Edit this page on GitHub'),
  locales: {
    root: { label: 'English', lang: 'en-US' },
    'zh-CN': { label: '简体中文', lang: 'zh-CN', themeConfig: localeTheme('指南', '/zh-CN/guide/introduction', zhSidebar, '在 GitHub 编辑此页') },
    'zh-TW': { label: '繁體中文', lang: 'zh-TW', themeConfig: localeTheme('指南', '/zh-TW/guide/introduction', zhTWSidebar, '在 GitHub 編輯此頁') },
    'ja-JP': { label: '日本語', lang: 'ja-JP', themeConfig: localeTheme('ガイド', '/ja-JP/guide/introduction', jaSidebar, 'GitHub でこのページを編集') },
    'ko-KR': { label: '한국어', lang: 'ko-KR', themeConfig: localeTheme('가이드', '/ko-KR/guide/introduction', koSidebar, 'GitHub에서 이 페이지 편집') }
  }
})
