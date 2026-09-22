export const githubRepository = 'buildwithtechx/outpipe';
export const githubRepositoryUrl = `https://github.com/${githubRepository}`;

interface GitHubRepositoryResponse {
  stargazers_count: number;
}

export async function getGitHubStarCount() {
  const response = await fetch(
    `https://api.github.com/repos/${githubRepository}`,
    {
      headers: {
        Accept: 'application/vnd.github+json',
      },
    },
  );

  if (!response.ok) {
    throw new Error('Unable to load the GitHub star count.');
  }

  const repository = (await response.json()) as GitHubRepositoryResponse;
  return repository.stargazers_count;
}

export function formatGitHubStarCount(count: number) {
  return new Intl.NumberFormat('en', {
    maximumFractionDigits: count >= 1000 ? 1 : 0,
    notation: 'compact',
  }).format(count);
}
