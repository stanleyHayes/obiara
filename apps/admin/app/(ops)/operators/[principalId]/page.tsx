import { OperatorsDesk } from "../operators-desk";

export default async function OperatorDetailPage({
  params,
}: Readonly<{ params: Promise<{ principalId: string }> }>) {
  const { principalId } = await params;
  return <OperatorsDesk key={principalId} principalId={principalId} />;
}
