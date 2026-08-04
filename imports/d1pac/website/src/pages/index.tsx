import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';
import styles from './index.module.css';

const steps = [
  ['01', 'Build', 'Compile ordinary SQLite SQL into a deterministic, reviewable package.'],
  ['02', 'Plan', 'Compare the package with the schema that D1 actually contains.'],
  ['03', 'Apply', 'Converge through a transactional D1 batch with explicit safety gates.'],
];

export default function Home() {
  return (
    <Layout
      title="Desired-state schemas for Cloudflare D1"
      description="Package, compare, and safely apply Cloudflare D1 schema state."
    >
      <main className={styles.page}>
        <section className={styles.hero}>
          <div className={`container ${styles.heroGrid}`}>
            <div>
              <p className={styles.eyebrow}>SQL package tooling for the edge</p>
              <Heading as="h1">Your D1 schema,<br /><span>declared.</span></Heading>
              <p className={styles.lead}>
                Keep the database you want in source control. d1pac packages that intent,
                compares it with live D1, and produces the exact path forward.
              </p>
              <div className={styles.actions}>
                <Link className={styles.primary} to="/manual/learn/quickstart">Start building</Link>
                <Link className={styles.secondary} to="/manual/reference/safety-model">Safety model</Link>
              </div>
            </div>
            <div className={styles.terminal}>
              <div className={styles.terminalBar}><span>deployment.plan</span><i>live D1</i></div>
              <div className={styles.planLine}><b>+</b> create table widget_events</div>
              <div className={styles.planLine}><em>~</em> alter table widgets add column status</div>
              <div className={styles.planLine}><b>+</b> create index idx_widgets_status</div>
              <footer><span>desired state</span><strong>3 operations</strong></footer>
            </div>
          </div>
        </section>
        <section className={styles.workflow}>
          <div className="container">
            <p className={styles.kicker}>The d1pac loop</p>
            <Heading as="h2">Build once. Compare reality. Converge safely.</Heading>
            <div className={styles.steps}>
              {steps.map(([number, title, description]) => (
                <article key={number}>
                  <span>{number}</span>
                  <Heading as="h3">{title}</Heading>
                  <p>{description}</p>
                  <code>d1pac {title.toLowerCase()}</code>
                </article>
              ))}
            </div>
          </div>
        </section>
        <section className={styles.detail}>
          <div className={`container ${styles.detailGrid}`}>
            <div>
              <p className={styles.kicker}>Built for SQLite semantics</p>
              <Heading as="h2">Declarative does not mean careless.</Heading>
            </div>
            <p>
              Additive changes stay small. Structural changes use SQLite's table-rebuild
              pattern and preserve common columns. Drops and removed columns remain blocked
              until the deployment explicitly authorizes them.
            </p>
          </div>
        </section>
      </main>
    </Layout>
  );
}
