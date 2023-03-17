use std::{env, fs::File, io::Write};

use clap::Parser;
use paper_mc::{request::DownloadRequired, PaperMcClient};

const BASE_URL: &str = "https://api.papermc.io";

#[derive(Debug, Parser)]
struct Args {
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
    let mut args = Args::parse();
    let url = env::var("PAPER_MC_BASE_URL").unwrap_or(BASE_URL.to_string());
    let client = PaperMcClient::new(&url);

    if args.projects {
        let projects_response = client
            .projects()
            .send()
            .await
            .expect("Failed to get projects");
        let projects = projects_response.projects.unwrap();
        return println!("Available Projects: {:?}", projects);
    }

    if args.version == "latest" {
        let project_response = client
            .project(&args.project)
            .send()
            .await
            .expect("Failed to get project");
        let versions = project_response.versions.unwrap();
        args.version = versions.last().unwrap().to_string();
    }

    if args.build <= 0 {
        let version_response = client
            .version(&args.project, &args.version)
            .send()
            .await
            .expect("Failed to get version");
        let builds = version_response.builds.unwrap();
        args.build = builds.last().unwrap().to_owned();
    }

    let build_response = client
        .build(&args.project, &args.version, args.build)
        .send()
        .await
        .expect("Failed to get build");
    let downloads = build_response.downloads.unwrap();

    if args.downloads {
        return println!("Available Downloads: {:#?}", downloads);
    }

    let download_name = downloads
        .get(&args.download)
        .unwrap()
        .get("name")
        .unwrap()
        .as_str()
        .unwrap();

    let download_required = DownloadRequired {
        project: &args.project,
        build: args.build,
        version: &args.version,
        download: download_name,
    };
    let download_response = client
        .download(download_required)
        .send()
        .await
        .expect("Failed to get download");
    let bytes = download_response.bytes().await.unwrap();

    if args.output == "source" {
        args.output = download_name.to_string();
    }

    let mut file = File::create(args.output).expect("Failed to create file");
    file.write_all(&bytes).expect("Failed to write file");
}
