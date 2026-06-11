import { Network } from "lucide-react";

import { ProjectsList } from "@/components/projects-list";

export default function App() {
  return (
    <main className="min-h-screen bg-background text-foreground">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-8 px-6 py-8 sm:px-8">
        <header className="flex flex-col gap-4 border-b pb-6 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-md border bg-card">
              <Network className="h-5 w-5" aria-hidden="true" />
            </div>
            <div>
              <h1 className="text-2xl font-semibold tracking-normal">
                Matryoshka
              </h1>
              <p className="text-sm text-muted-foreground">
                Project control plane
              </p>
            </div>
          </div>
        </header>

        <ProjectsList />
      </div>
    </main>
  );
}
