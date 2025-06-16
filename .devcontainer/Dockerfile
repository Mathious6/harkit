FROM mcr.microsoft.com/devcontainers/go:1.24

# This is important to change the user to clone/copy/etc. in good folder
USER vscode
WORKDIR /home/vscode

# Install Oh My Zsh : https://github.com/ohmyzsh/ohmyzsh?tab=readme-ov-file#basic-installation
# We need to use '|| true' to avoid the script to stop the execution if the installation is already done
RUN sh -c "$(wget https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh -O -)" || true

# Download and install ZSH plugins
RUN git clone https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions && \
    git clone https://github.com/zsh-users/zsh-syntax-highlighting.git ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting && \
    git clone https://github.com/MichaelAquilina/zsh-you-should-use.git ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/you-should-use

# Copy .devcontainer/zshrc config to apply plugins/themes/etc.
COPY --chown=vscode:vscode .zshrc .zshrc
