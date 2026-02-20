// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://zach-snell.github.io',
	base: '/bitbucket-go-mcp',
	integrations: [
		starlight({
			title: 'bitbucket-go-mcp',
			description: 'A Model Context Protocol (MCP) server for Bitbucket integration in Go.',
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/zach-snell/bitbucket-go-mcp' },
			],
			editLink: {
				baseUrl: 'https://github.com/zach-snell/bitbucket-go-mcp/edit/main/docs/',
			},
			customCss: ['./src/styles/custom.css'],
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'getting-started/introduction' },
						{ label: 'Installation', slug: 'getting-started/installation' },
						{ label: 'Configuration', slug: 'getting-started/configuration' },
					],
				},
			],
		}),
	],
});
