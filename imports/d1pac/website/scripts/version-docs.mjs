import {existsSync, readFileSync, writeFileSync} from 'node:fs';
import path from 'node:path';
import {spawnSync} from 'node:child_process';

const version = process.argv[2];
if (!version || !/^\d+\.\d+\.\d+$/.test(version)) {
  console.error('Usage: npm run version-docs -- X.Y.Z');
  process.exit(2);
}

const root = process.cwd();
const versionsPath = path.join(root, 'versions.json');
if (existsSync(versionsPath)) {
  const versions = JSON.parse(readFileSync(versionsPath, 'utf8'));
  if (Array.isArray(versions) && versions.includes(version)) {
    console.log(`Docs version ${version} already exists; skipping.`);
    process.exit(0);
  }
}

const binary = process.platform === 'win32'
  ? path.join(root, 'node_modules', '.bin', 'docusaurus.cmd')
  : path.join(root, 'node_modules', '.bin', 'docusaurus');
const result = spawnSync(binary, ['docs:version', version], {stdio: 'inherit'});
if (result.status !== 0) {
  process.exit(result.status ?? 1);
}

function section(markdown, header) {
  const start = markdown.indexOf(header);
  if (start === -1) return null;
  const remainder = markdown.slice(start + header.length);
  const next = remainder.search(/\n## \[/);
  return (next === -1 ? remainder : remainder.slice(0, next)).trim();
}

const changelog = readFileSync(path.resolve(root, '..', 'CHANGELOG.md'), 'utf8');
const released = section(changelog, `## [${version}]`);
const unreleased = section(changelog, '## [Unreleased]');
if (released === null || unreleased === null) {
  console.error('Expected release and Unreleased changelog sections.');
  process.exit(1);
}

const header = `---\nid: whats-new\ntitle: What's New\nslug: /whats-new\n---\n\n`;
const history = 'For the complete history, see [CHANGELOG.md](https://github.com/MagnusOpera/d1pac/blob/main/CHANGELOG.md).\n\n';
writeFileSync(path.join(root, 'versioned_docs', `version-${version}`, 'whats-new.md'), `${header}${history}## ${version}\n\n${released}\n`);
writeFileSync(path.join(root, 'docs', 'whats-new.md'), `${header}${history}## Unreleased\n\n${unreleased}\n`);
