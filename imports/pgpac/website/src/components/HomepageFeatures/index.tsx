import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  description: string;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Git is the timeline',
    description:
      'The SQL files are the desired schema. History, review, and rollback discussion stay in normal source control where teams already work.',
  },
  {
    title: 'The target database is the comparison point',
    description:
      'Plans are produced by comparing the packaged desired state with what PostgreSQL actually contains, not by assuming every migration ran perfectly.',
  },
  {
    title: 'Change execution stays explicit',
    description:
      'Build artifacts, inspect the computed plan, and apply with destructive-operation gates when the team intentionally allows them.',
  },
];

function Feature({title, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className={styles.featureCard}>
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props) => (
            <Feature key={props.title} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
