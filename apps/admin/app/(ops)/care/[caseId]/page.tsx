import { CareQueue } from "../care-queue";

export default async function CareCasePage({
  params,
}: Readonly<{ params: Promise<{ caseId: string }> }>) {
  const { caseId } = await params;
  return <CareQueue caseId={caseId} key={caseId} />;
}
