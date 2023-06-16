// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require("prism-react-renderer/themes/github");
const darkCodeTheme = require("prism-react-renderer/themes/oceanicNext");

/** @type {import('@docusaurus/types').Config} */
const config = {
	title: "Strolt",
	tagline: "Dinosaurs are cool",
	url: "https://strolt.shibanet0.com",
	baseUrl: "/",
	onBrokenLinks: "throw",
	onBrokenMarkdownLinks: "warn",
	favicon: "img/favicon.ico",

	// GitHub pages deployment config.
	// If you aren't using GitHub pages, you don't need these.
	organizationName: "strolt", // Usually your GitHub org/user name.
	projectName: "strolt", // Usually your repo name.
	deploymentBranch: 'gh-pages',
	trailingSlash: true,

	// Even if you don't use internalization, you can use this field to set useful
	// metadata like html lang. For example, if your site is Chinese, you may want
	// to replace "en" with "zh-Hans".
	i18n: {
		defaultLocale: "en",
		locales: ["en"],
	},

	presets: [
		[
			"classic",
			/** @type {import('@docusaurus/preset-classic').Options} */
			({
				docs: {
					sidebarPath: require.resolve("./sidebars.js"),

					editUrl: ({ locale, docPath }) => {
						if (locale !== "en") {
							return `https://github.com/strolt/strolt/edit/main/website/i18n/${locale}/docusaurus-plugin-content-docs/current/${docPath}`;
						}
						return `https://github.com/strolt/strolt/edit/main/website/docs/${docPath}`;
					},
				},
				theme: {
					customCss: require.resolve("./src/css/custom.css"),
				},
				gtag: {
          trackingID: 'G-1PKZLPTL7E',
          anonymizeIP: true,
        },
				sitemap: {
          changefreq: 'weekly',
          priority: 0.5,
          ignorePatterns: ['/tags/**'],
          filename: 'sitemap.xml',
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
						route: "/api/strolt",
					},
					{
						spec: "../.swagger/stroltm/swagger.yaml",
						route: "/api/stroltm",
					},
					{
						spec: "../.swagger/stroltp/swagger.yaml",
						route: "/api/stroltp",
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
			tableOfContents: {
				minHeadingLevel: 2,
				maxHeadingLevel: 5,
			},
			navbar: {
				title: "Strolt",
				logo: {
					alt: "My Site Logo",
					src: "img/logo.svg",
				},
				items: [
					{
						type: "doc",
						docId: "intro",
						position: "left",
						label: "Docs",
					},
					{ to: "/api/strolt", label: "API/Strolt", position: "left" },
					{ to: "/api/stroltm", label: "API/Stroltm", position: "left" },
					{ to: "/api/stroltp", label: "API/Stroltp", position: "left" },
					{
						href: "https://github.com/strolt/strolt",
						label: "GitHub",
						position: "right",
					},
					{
						type: "localeDropdown",
						position: "right",
					},
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
		}),
};

module.exports = config;
