use clap::Parser;
use paper_mc::{request::DownloadRequired, PaperMcClient};
use std::{fs::File, io::Write};

#[derive(Parser, Debug)]
struct Options {
    #[arg(short, long, default_value = "paper")]
    project: String,
    #[arg(long, default_value = "false", help = "List available projects")]
    projects: bool,
    #[arg(short, long, default_value = "latest", help = "Minecraft version")]
    version: String,
    #[arg(short, long, default_value = "-1", help = "Build of version")]
    build: i64,
    #[arg(short, long, default_value = "application")]
    download: String,
    #[arg(long, default_value = "false", help = "List available downloads")]
    downloads: bool,
    #[arg(short, long, default_value = "source", help = "Output file name")]
    output: String,
}

#[tokio::main]
async fn main() {
    let mut options = Options::parse();

    let url = std::env::var("PAPER_MC_BASE_URL").unwrap_or("https://api.papermc.io".to_string());
    let client = PaperMcClient::new(&url);

    if options.projects {
        let projects_response = client
            .projects()
            .send()
            .await
            .expect("Failed to get projects");
        let projects = projects_response.projects.unwrap();
        return println!("Available Projects: {:?}", projects);
    }

    if options.version == "latest" {
        let project_response = client
            .project(&options.project)
            .send()
            .await
            .expect("Failed to get project");
        let versions = project_response.versions.unwrap();
        options.version = versions.last().unwrap().to_string();
    }

    if options.build <= 0 {
        let version_response = client
            .version(&options.project, &options.version)
            .send()
            .await
            .expect("Failed to get version");
        let builds = version_response.builds.unwrap();
        options.build = builds.last().unwrap().to_owned();
    }

    let build_response = client
        .build(&options.project, &options.version, options.build)
        .send()
        .await
        .expect("Failed to get build");
    let downloads = build_response.downloads.unwrap();

    if options.downloads {
        return println!("Available Downloads: {:#?}", downloads);
    }

    let download_name = downloads
        .get(&options.download)
        .unwrap()
        .get("name")
        .unwrap()
        .as_str()
        .unwrap();

    let download_required = DownloadRequired {
        project: &options.project,
        build: options.build,
        version: &options.version,
        download: download_name,
    };
    let download_response = client
        .download(download_required)
        .send()
        .await
        .expect("Failed to get download");
    let bytes = download_response.bytes().await.unwrap();

    if options.output == "source" {
        options.output = download_name.to_string();
    }

    let mut file = File::create(options.output).expect("Failed to create file");
    file.write_all(&bytes).expect("Failed to write file");
}
