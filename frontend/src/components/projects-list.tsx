import { FormEvent, useEffect, useMemo, useState } from "react";
import { gql, useMutation, useQuery } from "@apollo/client";
import { ExternalLink, Plus, Rocket, RefreshCw } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

type Service = {
  id: string;
  projectId: string;
  name: string;
  type: string;
  status: string;
  externalPort: number | null;
};

type Project = {
  id: string;
  name: string;
  description: string;
};

type ProjectWithServices = Project & {
  services: Service[];
};

type ProjectsQuery = {
  projects: Project[];
};

type ProjectQuery = {
  project: ProjectWithServices | null;
};

type CreateProjectMutation = {
  createProject: Project;
};

type CreateProjectVariables = {
  input: {
    name: string;
    description: string;
  };
};

type CreateServiceMutation = {
  createService: Service;
};

type CreateServiceVariables = {
  input: {
    projectId: string;
    name: string;
    type: string;
  };
};

const PROJECTS_QUERY = gql`
  query Projects {
    projects {
      id
      name
      description
    }
  }
`;

const PROJECT_QUERY = gql`
  query Project($id: ID!) {
    project(id: $id) {
      id
      name
      description
      services {
        id
        projectId
        name
        type
        status
        externalPort
      }
    }
  }
`;

const CREATE_PROJECT_MUTATION = gql`
  mutation CreateProject($input: CreateProjectInput!) {
    createProject(input: $input) {
      id
      name
      description
    }
  }
`;

const CREATE_DATETIME_APP_MUTATION = gql`
  mutation CreateDatetimeApp($input: CreateServiceInput!) {
    createService(input: $input) {
      id
      projectId
      name
      type
      status
      externalPort
    }
  }
`;

export function ProjectsList() {
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(
    null
  );
  const [projectName, setProjectName] = useState("datetime-project");
  const [projectDescription, setProjectDescription] = useState(
    "Go datetime application environment"
  );
  const [serviceName, setServiceName] = useState("datetime-app");
  const [lastDeployedService, setLastDeployedService] =
    useState<Service | null>(null);

  const {
    data,
    loading,
    error,
    refetch: refetchProjects,
    networkStatus,
  } = useQuery<ProjectsQuery>(PROJECTS_QUERY, {
    notifyOnNetworkStatusChange: true,
  });

  const selectedProject = useMemo(
    () =>
      data?.projects.find((project) => project.id === selectedProjectId) ??
      data?.projects[0] ??
      null,
    [data?.projects, selectedProjectId]
  );

  const {
    data: projectData,
    loading: projectLoading,
    error: projectError,
    refetch: refetchProject,
  } = useQuery<ProjectQuery>(PROJECT_QUERY, {
    variables: { id: selectedProject?.id ?? "" },
    skip: !selectedProject,
  });

  const [createProject, { loading: creatingProject, error: createError }] =
    useMutation<CreateProjectMutation, CreateProjectVariables>(
      CREATE_PROJECT_MUTATION,
      {
        onCompleted: async ({ createProject }) => {
          setSelectedProjectId(createProject.id);
          setLastDeployedService(null);
          await refetchProjects();
        },
      }
    );

  const [createDatetimeApp, { loading: deployingApp, error: deployError }] =
    useMutation<CreateServiceMutation, CreateServiceVariables>(
      CREATE_DATETIME_APP_MUTATION,
      {
        onCompleted: async ({ createService }) => {
          setLastDeployedService(createService);
          await refetchProject();
        },
      }
    );

  useEffect(() => {
    if (!selectedProjectId && data?.projects[0]) {
      setSelectedProjectId(data.projects[0].id);
    }
  }, [data?.projects, selectedProjectId]);

  const isRefetching = networkStatus === 4;
  const services = projectData?.project?.services ?? [];

  async function handleCreateProject(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    await createProject({
      variables: {
        input: {
          name: projectName.trim(),
          description: projectDescription.trim(),
        },
      },
    });
  }

  async function handleDeployDatetimeApp(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!selectedProject) {
      return;
    }

    await createDatetimeApp({
      variables: {
        input: {
          projectId: selectedProject.id,
          name: serviceName.trim(),
          type: "app",
        },
      },
    });
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_380px]">
      <section className="space-y-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 className="text-lg font-semibold tracking-normal">Projects</h2>
            <p className="text-sm text-muted-foreground">
              Select a project, then deploy the Go datetime application.
            </p>
          </div>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => void refetchProjects()}
            disabled={loading || isRefetching}
          >
            <RefreshCw
              className={isRefetching ? "h-4 w-4 animate-spin" : "h-4 w-4"}
              aria-hidden="true"
            />
            Refresh
          </Button>
        </div>

        <ErrorAlert
          title="Unable to load projects"
          message={error?.message}
        />
        <ErrorAlert
          title="Unable to create project"
          message={createError?.message}
        />

        <div className="overflow-hidden rounded-md border bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Description</TableHead>
                <TableHead className="hidden w-[260px] xl:table-cell">
                  ID
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? <ProjectsSkeleton /> : null}
              {!loading && data?.projects.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={3}
                    className="h-28 text-center text-sm text-muted-foreground"
                  >
                    No projects yet. Create one to deploy the datetime app.
                  </TableCell>
                </TableRow>
              ) : null}
              {!loading
                ? data?.projects.map((project) => (
                    <TableRow
                      key={project.id}
                      className={
                        selectedProject?.id === project.id
                          ? "bg-muted/60"
                          : undefined
                      }
                    >
                      <TableCell>
                        <button
                          type="button"
                          className="text-left font-medium hover:underline"
                          onClick={() => {
                            setSelectedProjectId(project.id);
                            setLastDeployedService(null);
                          }}
                        >
                          {project.name}
                        </button>
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {project.description}
                      </TableCell>
                      <TableCell className="hidden font-mono text-xs text-muted-foreground xl:table-cell">
                        {project.id}
                      </TableCell>
                    </TableRow>
                  ))
                : null}
            </TableBody>
          </Table>
        </div>

        <ProjectServices
          loading={projectLoading}
          error={projectError?.message}
          project={projectData?.project ?? selectedProject}
          services={services}
          lastDeployedService={lastDeployedService}
        />
      </section>

      <aside className="space-y-4">
        <form
          onSubmit={(event) => void handleCreateProject(event)}
          className="space-y-4 rounded-md border bg-card p-4"
        >
          <div>
            <h2 className="text-base font-semibold tracking-normal">
              Create project
            </h2>
            <p className="text-sm text-muted-foreground">
              A project provisions its own Docker network.
            </p>
          </div>
          <div className="space-y-2">
            <Label htmlFor="project-name">Name</Label>
            <Input
              id="project-name"
              value={projectName}
              onChange={(event) => setProjectName(event.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="project-description">Description</Label>
            <Input
              id="project-description"
              value={projectDescription}
              onChange={(event) => setProjectDescription(event.target.value)}
              required
            />
          </div>
          <Button
            type="submit"
            className="w-full"
            disabled={
              creatingProject ||
              projectName.trim() === "" ||
              projectDescription.trim() === ""
            }
          >
            <Plus aria-hidden="true" />
            {creatingProject ? "Creating..." : "Create project"}
          </Button>
        </form>

        <form
          onSubmit={(event) => void handleDeployDatetimeApp(event)}
          className="space-y-4 rounded-md border bg-card p-4"
        >
          <div>
            <h2 className="text-base font-semibold tracking-normal">
              Deploy datetime app
            </h2>
            <p className="text-sm text-muted-foreground">
              Launches the bundled Go service that returns the current time.
            </p>
          </div>
          <ErrorAlert
            title="Unable to deploy datetime app"
            message={deployError?.message}
          />
          <div className="space-y-2">
            <Label htmlFor="selected-project">Target project</Label>
            <Input
              id="selected-project"
              value={selectedProject?.name ?? "No project selected"}
              readOnly
              disabled={!selectedProject}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="service-name">Service name</Label>
            <Input
              id="service-name"
              value={serviceName}
              onChange={(event) => setServiceName(event.target.value)}
              required
            />
          </div>
          <Button
            type="submit"
            className="w-full"
            disabled={
              !selectedProject || deployingApp || serviceName.trim() === ""
            }
          >
            <Rocket aria-hidden="true" />
            {deployingApp ? "Deploying..." : "Deploy datetime app"}
          </Button>
        </form>
      </aside>
    </div>
  );
}

function ProjectServices({
  loading,
  error,
  project,
  services,
  lastDeployedService,
}: {
  loading: boolean;
  error?: string;
  project: Project | ProjectWithServices | null;
  services: Service[];
  lastDeployedService: Service | null;
}) {
  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold tracking-normal">
          {project ? project.name : "Project services"}
        </h2>
        <p className="text-sm text-muted-foreground">
          {project
            ? "Services deployed into this project network."
            : "Select a project to inspect services."}
        </p>
      </div>

      <ErrorAlert title="Unable to load services" message={error} />

      {lastDeployedService?.externalPort ? (
        <Alert>
          <AlertTitle>Datetime app deployed</AlertTitle>
          <AlertDescription>
            <a
              className="inline-flex items-center gap-1 font-medium underline-offset-4 hover:underline"
              href={`http://localhost:${lastDeployedService.externalPort}/`}
              target="_blank"
              rel="noreferrer"
            >
              Open http://localhost:{lastDeployedService.externalPort}/
              <ExternalLink className="h-3.5 w-3.5" aria-hidden="true" />
            </a>
          </AlertDescription>
        </Alert>
      ) : null}

      <div className="overflow-hidden rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Public URL</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? <ServicesSkeleton /> : null}
            {!loading && project && services.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={4}
                  className="h-24 text-center text-sm text-muted-foreground"
                >
                  No services deployed yet.
                </TableCell>
              </TableRow>
            ) : null}
            {!loading && !project ? (
              <TableRow>
                <TableCell
                  colSpan={4}
                  className="h-24 text-center text-sm text-muted-foreground"
                >
                  No project selected.
                </TableCell>
              </TableRow>
            ) : null}
            {!loading
              ? services.map((service) => (
                  <TableRow key={service.id}>
                    <TableCell className="font-medium">
                      {service.name}
                    </TableCell>
                    <TableCell>{service.type}</TableCell>
                    <TableCell>
                      <ServiceStatus status={service.status} />
                    </TableCell>
                    <TableCell>
                      {service.externalPort ? (
                        <a
                          href={`http://localhost:${service.externalPort}/`}
                          target="_blank"
                          rel="noreferrer"
                          className="inline-flex items-center gap-1 text-sm font-medium underline-offset-4 hover:underline"
                        >
                          localhost:{service.externalPort}
                          <ExternalLink
                            className="h-3.5 w-3.5"
                            aria-hidden="true"
                          />
                        </a>
                      ) : (
                        <span className="text-sm text-muted-foreground">
                          Internal only
                        </span>
                      )}
                    </TableCell>
                  </TableRow>
                ))
              : null}
          </TableBody>
        </Table>
      </div>
    </section>
  );
}

function ServiceStatus({ status }: { status: string }) {
  if (status === "running") {
    return <Badge variant="default">Running</Badge>;
  }

  if (status === "failed") {
    return <Badge variant="destructive">Failed</Badge>;
  }

  return <Badge variant="secondary">{status}</Badge>;
}

function ErrorAlert({ title, message }: { title: string; message?: string }) {
  if (!message) {
    return null;
  }

  return (
    <Alert variant="destructive">
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription>{message}</AlertDescription>
    </Alert>
  );
}

function ProjectsSkeleton() {
  return Array.from({ length: 4 }).map((_, index) => (
    <TableRow key={index}>
      <TableCell>
        <Skeleton className="h-4 w-36" />
      </TableCell>
      <TableCell>
        <Skeleton className="h-4 w-full max-w-md" />
      </TableCell>
      <TableCell className="hidden xl:table-cell">
        <Skeleton className="h-4 w-56" />
      </TableCell>
    </TableRow>
  ));
}

function ServicesSkeleton() {
  return Array.from({ length: 3 }).map((_, index) => (
    <TableRow key={index}>
      <TableCell>
        <Skeleton className="h-4 w-32" />
      </TableCell>
      <TableCell>
        <Skeleton className="h-4 w-16" />
      </TableCell>
      <TableCell>
        <Skeleton className="h-5 w-20" />
      </TableCell>
      <TableCell>
        <Skeleton className="h-4 w-28" />
      </TableCell>
    </TableRow>
  ));
}
