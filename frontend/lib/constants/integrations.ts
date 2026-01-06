export const IntegrationType = {
  CODE: "code",
  PROFESSIONAL: "professional",
  EDUCATION: "education",
  CERTIFICATION: "certification",
  SOCIAL: "social",
} as const;

export type IntegrationType =
  (typeof IntegrationType)[keyof typeof IntegrationType];

export interface IntegrationMetadata {
  id: string;
  name: string;
  type: IntegrationType;
  description: string;
  icon: string;
  color: string;
  comingSoon?: boolean;
}

export const integrations: IntegrationMetadata[] = [
  // Code
  {
    id: "github",
    name: "GitHub",
    type: IntegrationType.CODE,
    description: "Verify commits, repos, and contributions",
    icon: "github",
    color: "#181717",
  },
  {
    id: "gitlab",
    name: "GitLab",
    type: IntegrationType.CODE,
    description: "Verify projects and merge requests",
    icon: "gitlab",
    color: "#FC6D26",
    comingSoon: true,
  },
  {
    id: "stackoverflow",
    name: "Stack Overflow",
    type: IntegrationType.CODE,
    description: "Verify reputation and contributions",
    icon: "message-circle-question",
    color: "#F58025",
    comingSoon: true,
  },

  // Professional
  {
    id: "linkedin",
    name: "LinkedIn",
    type: IntegrationType.PROFESSIONAL,
    description: "Verify employment and connections",
    icon: "linkedin",
    color: "#0A66C2",
  },
  {
    id: "upwork",
    name: "Upwork",
    type: IntegrationType.PROFESSIONAL,
    description: "Verify freelance history and ratings",
    icon: "briefcase",
    color: "#14A800",
    comingSoon: true,
  },

  // Education
  {
    id: "coursera",
    name: "Coursera",
    type: IntegrationType.EDUCATION,
    description: "Verify course completions and certificates",
    icon: "graduation-cap",
    color: "#0056D2",
    comingSoon: true,
  },
  {
    id: "udemy",
    name: "Udemy",
    type: IntegrationType.EDUCATION,
    description: "Verify completed courses",
    icon: "book-open",
    color: "#A435F0",
    comingSoon: true,
  },

  // Certifications
  {
    id: "aws",
    name: "AWS",
    type: IntegrationType.CERTIFICATION,
    description: "Verify cloud certifications",
    icon: "cloud",
    color: "#FF9900",
    comingSoon: true,
  },
  {
    id: "google-cloud",
    name: "Google Cloud",
    type: IntegrationType.CERTIFICATION,
    description: "Verify GCP certifications",
    icon: "cloud",
    color: "#4285F4",
    comingSoon: true,
  },
];

export function getIntegration(id: string): IntegrationMetadata | undefined {
  return integrations.find((i) => i.id === id);
}

export function getIntegrationsByType(
  type: IntegrationType
): IntegrationMetadata[] {
  return integrations.filter((i) => i.type === type);
}