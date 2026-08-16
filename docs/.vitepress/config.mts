import { defineConfig } from 'vitepress'

const repo = 'https://github.com/mouse233/MySekaiMapper'

const enSidebar = [
  {
    text: 'Guide',
    items: [
      { text: 'Introduction', link: '/guide/introduction' },
      { text: 'Quick Start', link: '/guide/quickstart' },
      { text: 'Upload API', link: '/guide/upload-api' },
      { text: 'Push Mechanism', link: '/guide/push' },
      { text: 'Static File Server', link: '/guide/static-server' },
      { text: 'Player Routing', link: '/guide/routing' },
      { text: 'CLI Reference', link: '/guide/cli' }
    ]
  },
  {
    text: 'Project',
    items: [
      { text: 'Directory Structure', link: '/reference/structure' },
      { text: 'FAQ', link: '/faq' },
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
      { text: '上传接口', link: '/zh-CN/guide/upload-api' },
      { text: '推送机制', link: '/zh-CN/guide/push' },
      { text: '静态文件服务器', link: '/zh-CN/guide/static-server' },
      { text: '玩家推送路由', link: '/zh-CN/guide/routing' },
      { text: '命令行工具', link: '/zh-CN/guide/cli' }
    ]
  },
  {
    text: '项目',
    items: [
      { text: '目录结构', link: '/zh-CN/reference/structure' },
      { text: '常见问题', link: '/zh-CN/faq' },
      { text: '许可证', link: '/zh-CN/license' }
    ]
  }
]

export default defineConfig({
  // GitHub Pages 项目站点必须设置 base
  base: '/MySekaiMapper/',
  cleanUrls: true,
  lastUpdated: true,
  markdown: {
    lineNumbers: true
  },
  locales: {
    root: {
      label: 'English',
      lang: 'en-US',
      title: 'MySekaiMapper',
      description:
        'A resource-gathering point map generator and auto-notifier for MySekai in Project SEKAI',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/guide/introduction', activeMatch: '/guide/' },
          { text: 'FAQ', link: '/faq' },
          { text: 'License', link: '/license' },
          { text: 'GitHub', link: repo }
        ],
        sidebar: enSidebar,
        outline: { label: 'On this page', level: [2, 3] },
        docFooter: { prev: 'Previous page', next: 'Next page' },
        lastUpdated: { text: 'Last updated' },
        returnToTopLabel: 'Back to top',
        sidebarMenuLabel: 'Menu',
        darkModeSwitchLabel: 'Appearance',
        lightModeSwitchTitle: 'Switch to light mode',
        darkModeSwitchTitle: 'Switch to dark mode',
        editLink: {
          pattern: `${repo}/edit/main/docs/:path`,
          text: 'Edit this page on GitHub'
        },
        socialLinks: [{ icon: 'github', link: repo }]
      }
    },
    'zh-CN': {
      label: '简体中文',
      lang: 'zh-CN',
      title: 'MySekaiMapper',
      description: '《世界计划 多彩舞台》MySekai 采集点地图生成与自动推送工具',
      themeConfig: {
        nav: [
          { text: '指南', link: '/zh-CN/guide/introduction', activeMatch: '/zh-CN/guide/' },
          { text: '常见问题', link: '/zh-CN/faq' },
          { text: '许可证', link: '/zh-CN/license' },
          { text: 'GitHub', link: repo }
        ],
        sidebar: zhSidebar,
        outline: { label: '本页目录', level: [2, 3] },
        docFooter: { prev: '上一页', next: '下一页' },
        lastUpdated: { text: '最后更新' },
        returnToTopLabel: '回到顶部',
        sidebarMenuLabel: '菜单',
        darkModeSwitchLabel: '外观',
        lightModeSwitchTitle: '切换到浅色模式',
        darkModeSwitchTitle: '切换到深色模式',
        editLink: {
          pattern: `${repo}/edit/main/docs/:path`,
          text: '在 GitHub 上编辑此页'
        },
        socialLinks: [{ icon: 'github', link: repo }]
      }
    }
  }
})
