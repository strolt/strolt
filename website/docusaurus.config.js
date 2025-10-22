// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require("prism-react-renderer/themes/github");
const darkCodeTheme = require("prism-react-renderer/themes/oceanicNext");
const { baseURL } = require("./config");

/** @type {import('@docusaurus/types').Config} */
const config = {
	title: "strolt",
	tagline: "A user-friendly tool for the effortless backup management.",
	url: baseURL,
	baseUrl: "/",
	onBrokenLinks: "throw",
	onBrokenMarkdownLinks: "warn",
	favicon: "img/favicon.svg",

	// GitHub pages deployment config.
	// If you aren't using GitHub pages, you don't need these.
	organizationName: "strolt", // Usually your GitHub org/user name.
	projectName: "strolt", // Usually your repo name.
	deploymentBranch: "gh-pages",

	// Even if you don't use internalization, you can use this field to set useful
	// metadata like html lang. For example, if your site is Chinese, you may want
	// to replace "en" with "zh-Hans".
	i18n: {
		defaultLocale: "en",
		locales: ["en"],
	},

	scripts: [
		process.env.NODE_ENV === "production" && {
			src: "https://a.shibanet0.com/pzjlkgj6ujcurpo",
			"data-website-id": "dc5cd938-935a-4b1e-95f8-da4d89f043ac",
			async: true,
		},
	].filter(Boolean),

	presets: [
		[
			"classic",
			/** @type {import('@docusaurus/preset-classic').Options} */
			({
				docs: {
					sidebarPath: require.resolve("./sidebars.js"),

					editUrl: ({ locale, docPath }) => {
						// if (locale !== "en") {
						// 	return `https://github.com/strolt/strolt/edit/main/website/i18n/${locale}/docusaurus-plugin-content-docs/current/${docPath}`;
						// }
						return `https://github.com/strolt/strolt/edit/main/website/docs/${docPath}`;
					},
				},
				theme: {
					customCss: require.resolve("./src/css/custom.css"),
				},
				sitemap: {
					changefreq: "weekly",
					priority: 0.5,
					ignorePatterns: ["/tags/**"],
					filename: "sitemap.xml",
				},
			}),
		],
		[
			"redocusaurus",
			{
				// Plugin Options for loading OpenAPI files
				specs: [
					{
						spec: "../.swagger/strolt/swagger.yaml",
						route: "/docs/api/strolt",
					},
					{
						spec: "../.swagger/stroltm/swagger.yaml",
						route: "/docs/api/stroltm",
					},
					{
						spec: "../.swagger/stroltp/swagger.yaml",
						route: "/docs/api/stroltp",
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

	themeConfig:
		/** @type {import('@docusaurus/preset-classic').ThemeConfig} */
		({
			metadata: [
				{
					name: "keywords",
					content: "strolt, backup, restic, pg_dump, mongodump, mariadb-dump",
				},
				// { name: "og:image", content: `${baseURL}/img/logo.svg` },
				// { name: "twitter:image", content: `${baseURL}/img/logo.svg` },
			],
			tableOfContents: {
				minHeadingLevel: 2,
				maxHeadingLevel: 5,
			},
			navbar: {
				title: "Strolt",
				logo: {
					alt: "strolt - logo",
					src: "img/favicon.svg",
				},
				items: [
					{
						type: "doc",
						docId: "intro",
						position: "left",
						label: "Docs",
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
			},
			footer: {
				style: "dark",
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
			},
			prism: {
				theme: lightCodeTheme,
				darkTheme: darkCodeTheme,
			},
			colorMode: {
				respectPrefersColorScheme: true,
			},
			algolia: {
				appId: "FF83SZ9CDS",
				apiKey: "643d73c08ff386e3da20a56e59319785", // pragma: allowlist secret
				indexName: "strolt-shibanet0",
				contextualSearch: true,
			},
		}),
};

module.exports = config;
