// Shape of one evidence metric rendered by the analytics dashboard. Metrics
// are composed from the live funnel report in analytics-dashboard.tsx; this
// module intentionally holds no seeded or fabricated values.
export interface GateMetric {
  readonly id: string;
  readonly label: string;
  readonly numerator: number;
  readonly denominator: number;
  readonly threshold: string;
  readonly value: string;
  readonly complete: boolean;
  readonly passes: boolean;
}
