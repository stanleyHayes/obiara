import { VerificationQueue } from "../verification-queue";

export default async function VerificationCasePage({
  params,
}: Readonly<{ params: Promise<{ caseId: string }> }>) {
  const { caseId } = await params;
  return <VerificationQueue caseId={caseId} key={caseId} />;
}
