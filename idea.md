# Yumdown - Music downloader from video

Desktop app to download audio from video url

## Features
- Input to URL from video (required)
- Input to filename (optional)
- Select folder to save audio
- Select output audio format
- Support multiple URLs and download simultaneously
- Open output folder
- Advanced tab for advanced features
- Keep the history in another tab (url, output folder, datetime, status: ok/error, audio output filesize)
- Show progress bar when dowloading (per each and total)

## Tech
- Use https://github.com/yt-dlp/yt-dlp to download audio from video url
- Use Wails (https://wails.io/) with React (Typescript, Vite, TailwindCSS, PNPM, Prettier and ESLint)
- Use the correct yt-dlp executable/binary by operating system
- Follow style guide from base-ui.com (https://base-ui.com/react/overview/quick-start)
- Frontend should be use atomic design
