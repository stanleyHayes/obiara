import { SafetyDesk } from "../safety-desk";

export default async function SafetyCasePage({
  params,
}: Readonly<{ params: Promise<{ caseId: string }> }>) {
  const { caseId } = await params;
  return <SafetyDesk caseId={caseId} key={caseId} />;
}
