import {existsSync, readFileSync, writeFileSync} from 'node:fs';
import path from 'node:path';
import {spawnSync} from 'node:child_process';

const [tool, version] = process.argv.slice(2);
if (!['pgpac', 'd1pac'].includes(tool) || !/^\d+\.\d+\.\d+$/.test(version ?? '')) {
  console.error('Usage: npm run version-docs -- <pgpac|d1pac> X.Y.Z');
  process.exit(2);
}

const root = process.cwd();
const versionsPath = path.join(root, `${tool}_versions.json`);

function extractSection(markdown, header) {
  const start = markdown.indexOf(header);
  if (start === -1) {
    return null;
  }

  const afterHeader = markdown.slice(start + header.length);
  const nextHeaderOffset = afterHeader.search(/\n## \[/);
  return (nextHeaderOffset === -1 ? afterHeader : afterHeader.slice(0, nextHeaderOffset)).trim();
}

if (existsSync(versionsPath)) {
  const versions = JSON.parse(readFileSync(versionsPath, 'utf8'));
  if (Array.isArray(versions) && versions.includes(version)) {
    console.log(`${tool} docs version ${version} already exists; skipping.`);
    process.exit(0);
  }
}

const bin = process.platform === 'win32'
  ? path.join(root, 'node_modules', '.bin', 'docusaurus.cmd')
  : path.join(root, 'node_modules', '.bin', 'docusaurus');
const result = spawnSync(bin, [`docs:version:${tool}`, version], {stdio: 'inherit'});
if (result.status !== 0) {
  process.exit(result.status ?? 1);
}

const changelogPath = path.resolve(root, '..', 'products', tool, 'CHANGELOG.md');
const changelog = readFileSync(changelogPath, 'utf8');
const releasedSectionBody = extractSection(changelog, `## [${version}]`);
const unreleasedSectionBody = extractSection(changelog, '## [Unreleased]');
if (releasedSectionBody === null || unreleasedSectionBody === null) {
  console.error(`Missing ${tool} changelog section for ${version} or Unreleased.`);
  process.exit(1);
}

const frontmatter = `---
id: whats-new
title: What's New
slug: /whats-new
---`;
const changelogUrl = `https://github.com/MagnusOpera/starpac/blob/main/products/${tool}/CHANGELOG.md`;
const versionedWhatsNew = path.join(root, `${tool}_versioned_docs`, `version-${version}`, 'whats-new.md');
writeFileSync(versionedWhatsNew, `${frontmatter}\n\nFor the complete history, see the full [CHANGELOG.md](${changelogUrl}) on GitHub.\n\n## ${version}\n\n${releasedSectionBody}\n`);

const currentWhatsNew = path.join(root, 'docs', tool, 'whats-new.md');
writeFileSync(currentWhatsNew, `${frontmatter}\n\nFor the complete history, see the full [CHANGELOG.md](${changelogUrl}) on GitHub.\n\n## Unreleased\n\n${unreleasedSectionBody}\n`);
