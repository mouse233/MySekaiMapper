import { defineConfig } from 'vitepress'

const repo = 'https://github.com/mouse233/MySekaiMapper'

const enSidebar = [
  {
    text: 'Go Guide',
    items: [
      { text: 'Quick Start', link: '/guide/quickstart' },
      { text: 'CLI Reference', link: '/guide/cli' },
      { text: 'Upload API', link: '/guide/upload-api' },
      { text: 'Reqable Report Server', link: '/guide/report-server' },
      { text: 'Notifications', link: '/guide/notifications' },
      { text: 'Go Refactor', link: '/guide/refactor-go' }
    ]
  },
  { text: 'Project', items: [{ text: 'Directory Structure', link: '/reference/structure' }] }
]

const zhSidebar = [
  {
    text: 'Go 指南',
    items: [
      { text: '快速上手', link: '/zh-CN/guide/quickstart' },
      { text: '命令行参考', link: '/zh-CN/guide/cli' },
      { text: '上传接口', link: '/zh-CN/guide/upload-api' },
      { text: 'Reqable 上报服务器', link: '/zh-CN/guide/report-server' },
      { text: '通知与路由', link: '/zh-CN/guide/notifications' },
      { text: 'Go 重构说明', link: '/zh-CN/guide/refactor-go' }
    ]
  },
  { text: '项目', items: [{ text: '目录结构', link: '/zh-CN/reference/structure' }] }
]

export default defineConfig({
  title: 'MySekaiMapper',
  description: 'A Go resource-gathering point map generator for MySekai',
  cleanUrls: true,
  themeConfig: {
    nav: [
      { text: 'Guide', link: '/guide/quickstart' },
      { text: 'GitHub', link: repo }
    ],
    sidebar: enSidebar,
    socialLinks: [{ icon: 'github', link: repo }],
    editLink: { pattern: `${repo}/edit/main/docs/:path`, text: 'Edit this page on GitHub' }
  },
  locales: {
    root: { label: 'English', lang: 'en-US' },
    'zh-CN': {
      label: '简体中文',
      lang: 'zh-CN',
      themeConfig: {
        nav: [
          { text: '指南', link: '/zh-CN/guide/quickstart' },
          { text: 'GitHub', link: repo }
        ],
        sidebar: zhSidebar,
        socialLinks: [{ icon: 'github', link: repo }],
        editLink: { pattern: `${repo}/edit/main/docs/:path`, text: '在 GitHub 编辑此页' }
      }
    }
  }
})
