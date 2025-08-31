# LangChain Go 中文版文档网站设置

## 选项 1: MkDocs (推荐)

### 安装
```bash
pip install mkdocs mkdocs-material
```

### 初始化
```bash
mkdocs new .
```

### 配置文件 (mkdocs.yml)
```yaml
site_name: LangChain Go 中文版
site_description: LangChain Go 语言实现的中文文档
site_url: https://your-username.github.io/langchaingo-cn

theme:
  name: material
  language: zh
  features:
    - navigation.tabs
    - navigation.sections
    - navigation.expand
    - search.highlight
    - search.share
  palette:
    - scheme: default
      primary: indigo
      accent: indigo
      toggle:
        icon: material/brightness-7
        name: Switch to dark mode
    - scheme: slate
      primary: indigo
      accent: indigo
      toggle:
        icon: material/brightness-4
        name: Switch to light mode

nav:
  - 首页: index.md
  - 快速开始: getting-started.md
  - 配置使用: config_usage.md
  - API 参考:
    - LLMs: api/llms.md
    - Schema: api/schema.md
    - Graph: api/graph.md
  - 示例: examples.md

plugins:
  - search:
      lang: zh
  - awesome-pages

markdown_extensions:
  - codehilite
  - admonition
  - toc:
      permalink: true
```

### 运行
```bash
mkdocs serve  # 开发模式
mkdocs build  # 构建静态文件
```

## 选项 2: Docusaurus

### 安装
```bash
npx create-docusaurus@latest website classic
```

### 配置文件 (docusaurus.config.js)
```javascript
module.exports = {
  title: 'LangChain Go 中文版',
  tagline: 'LangChain Go 语言实现的中文文档',
  url: 'https://your-username.github.io',
  baseUrl: '/langchaingo-cn/',
  
  i18n: {
    defaultLocale: 'zh-CN',
    locales: ['zh-CN'],
  },
  
  themeConfig: {
    navbar: {
      title: 'LangChain Go',
      items: [
        {to: '/docs/intro', label: '文档', position: 'left'},
        {to: '/docs/config_usage', label: '配置使用', position: 'left'},
        {
          href: 'https://github.com/sjzsdu/langchaingo-cn',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
  },
  
  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};
```

## 选项 3: GitHub Pages + Jekyll

### _config.yml
```yaml
title: LangChain Go 中文版
description: LangChain Go 语言实现的中文文档
baseurl: "/langchaingo-cn"
url: "https://your-username.github.io"

markdown: kramdown
highlighter: rouge
theme: minima

plugins:
  - jekyll-feed
  - jekyll-sitemap

collections:
  docs:
    output: true
    permalink: /:collection/:name/
```

## 部署选项

### 1. GitHub Pages
```yaml
# .github/workflows/docs.yml
name: Deploy docs
on:
  push:
    branches: [ main ]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Setup Python
      uses: actions/setup-python@v2
      with:
        python-version: '3.x'
    - name: Install dependencies
      run: |
        pip install mkdocs mkdocs-material
    - name: Build docs
      run: mkdocs build
    - name: Deploy to GitHub Pages
      uses: peaceiris/actions-gh-pages@v3
      with:
        github_token: ${{ secrets.GITHUB_TOKEN }}
        publish_dir: ./site
```

### 2. Netlify
- 直接连接 GitHub 仓库
- 自动部署
- 支持自定义域名

### 3. Vercel
- 零配置部署
- 全球 CDN
- 支持 serverless 函数

## 推荐流程

1. **选择 MkDocs Material 主题** - 最适合技术文档
2. **设置 GitHub Actions** - 自动部署到 GitHub Pages
3. **组织文档结构**：
   ```
   docs/
   ├── index.md          # 首页
   ├── getting-started.md # 快速开始
   ├── config_usage.md   # 配置使用 (已有)
   ├── api/              # API 文档
   │   ├── llms.md
   │   ├── schema.md
   │   └── graph.md
   └── examples/         # 示例
       ├── basic.md
       └── advanced.md
   ```

要我帮你设置其中任何一个方案吗？
