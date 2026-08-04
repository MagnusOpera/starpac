import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const baseUrl = '/d1pac/';

const config: Config = {
  title: 'd1pac',
  tagline: 'Desired-state schema deployment for Cloudflare D1.',
  favicon: 'img/logo.svg',
  future: {v4: true},
  url: 'https://magnusopera.github.io',
  baseUrl,
  organizationName: 'MagnusOpera',
  projectName: 'd1pac',
  onBrokenLinks: 'throw',
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },
  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          routeBasePath: 'manual',
          lastVersion: process.env.D1PAC_DOCS_LAST_VERSION ?? 'current',
          versions: {
            current: {label: 'Next'},
          },
          editUrl: 'https://github.com/MagnusOpera/d1pac/tree/main/website/',
        },
        blog: false,
        theme: {customCss: './src/css/custom.css'},
      } satisfies Preset.Options,
    ],
  ],
  themeConfig: {
    image: 'img/logo.svg',
    navbar: {
      title: 'd1pac',
      logo: {alt: 'd1pac logo', src: 'img/logo.svg'},
      items: [
        {type: 'docSidebar', sidebarId: 'docs', position: 'left', label: 'Docs'},
        {type: 'docsVersionDropdown', position: 'left'},
        {href: 'https://github.com/MagnusOpera/d1pac', label: 'GitHub', position: 'right'},
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Learn',
          items: [
            {label: 'Installation', to: '/manual/learn/installation'},
            {label: 'Quickstart', to: '/manual/learn/quickstart'},
          ],
        },
        {
          title: 'Reference',
          items: [
            {label: 'Project file', to: '/manual/reference/project-file'},
            {label: 'Safety model', to: '/manual/reference/safety-model'},
          ],
        },
        {
          title: 'Project',
          items: [
            {label: 'Repository', href: 'https://github.com/MagnusOpera/d1pac'},
            {label: 'Changelog', href: 'https://github.com/MagnusOpera/d1pac/blob/main/CHANGELOG.md'},
          ],
        },
      ],
      copyright: `© ${new Date().getFullYear()} Magnus Opera SAS`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['sql', 'bash', 'json'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
