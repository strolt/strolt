import type * as Preset from "@docusaurus/preset-classic";
import type { Config } from "@docusaurus/types";

import { themes as prismThemes } from "prism-react-renderer";

import { baseURL } from "./config";

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config: Config = {
  baseUrl: "/",
  deploymentBranch: "gh-pages",
  favicon: "img/favicon.svg",
  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: {
    v4: true, // Improve compatibility with the upcoming Docusaurus v4
  },
  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },
  markdown: {
    hooks: {
      onBrokenMarkdownLinks: "warn",
    },
  },
  onBrokenLinks: "throw",

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: "strolt", // Usually your GitHub org/user name.

  presets: [
    [
      "classic",
      {
        docs: {
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl: ({ docPath, locale }) => {
            // if (locale !== "en") {
            // 	return `https://github.com/strolt/strolt/edit/main/website/i18n/${locale}/docusaurus-plugin-content-docs/current/${docPath}`;
            // }
            return `https://github.com/strolt/strolt/edit/main/website/docs/${docPath}`;
          },
          sidebarPath: "./sidebars.ts",
        },
        sitemap: {
          changefreq: "weekly",
          filename: "sitemap.xml",
          ignorePatterns: ["/tags/**"],
          priority: 0.5,
        },
        theme: {
          customCss: "./src/css/custom.css",
        },
      } satisfies Preset.Options,
    ],
    [
      "redocusaurus",
      {
        // Plugin Options for loading OpenAPI files
        specs: [
          {
            route: "/docs/api/strolt",
            spec: "../.swagger/strolt/swagger.yaml",
          },
          {
            route: "/docs/api/stroltm",
            spec: "../.swagger/stroltm/swagger.yaml",
          },
          {
            route: "/docs/api/stroltp",
            spec: "../.swagger/stroltp/swagger.yaml",
          },
        ],
        // Theme Options for modifying how redoc renders them
        theme: {
          // Change with your site colors
          primaryColor: "#1890ff",
        },
      },
    ],
  ],
  projectName: "strolt", // Usually your repo name.
  scripts: [
    process.env.NODE_ENV === "production" && {
      async: true,
      "data-website-id": "dc5cd938-935a-4b1e-95f8-da4d89f043ac",
      src: "https://a.shibanet0.com/pzjlkgj6ujcurpo",
    },
  ].filter(Boolean),

  tagline: "A user-friendly tool for the effortless backup management.",

  themeConfig: {
    algolia: {
      apiKey: "643d73c08ff386e3da20a56e59319785", // pragma: allowlist secret
      appId: "FF83SZ9CDS",
      contextualSearch: true,
      indexName: "strolt-shibanet0",
    },
    colorMode: {
      respectPrefersColorScheme: true,
    },
    footer: {
      // links: [
      //   {
      //     title: 'Docs',
      //     items: [
      //       {
      //         label: 'Docs',
      //         to: '/docs/intro',
      //       },
      //     ],
      //   },
      //   {
      //     title: 'More',
      //     items: [
      //       {
      //         label: 'GitHub',
      //         href: 'https://github.com/strolt/strolt',
      //       },
      //     ],
      //   },
      // ],
      copyright: `Copyright © ${new Date().getFullYear()} Strolt. Built with Docusaurus.`,
      style: "dark",
    },
    metadata: [
      {
        content: "strolt, backup, restic, pg_dump, mongodump, mysqldump",
        name: "keywords",
      },
      // { name: "og:image", content: `${baseURL}/img/logo.svg` },
      // { name: "twitter:image", content: `${baseURL}/img/logo.svg` },
    ],
    navbar: {
      items: [
        {
          docId: "intro",
          label: "Docs",
          position: "left",
          type: "doc",
        },
        {
          href: "https://github.com/strolt/strolt",
          label: "GitHub",
          position: "right",
        },
        // {
        // 	type: "localeDropdown",
        // 	position: "right",
        // },
      ],
      logo: {
        alt: "strolt - logo",
        src: "img/favicon.svg",
      },
      title: "Strolt",
    },
    prism: {
      darkTheme: prismThemes.oceanicNext,
      theme: prismThemes.github,
    },
    tableOfContents: {
      maxHeadingLevel: 5,
      minHeadingLevel: 2,
    },
  } satisfies Preset.ThemeConfig,

  title: "strolt",

  url: baseURL,
};

export default config;
