#!/bin/bash
set -e

echo "Starting Manim Renderer..."
echo "Quality: ${MANIM_QUALITY:-medium}"
echo "Color: ${MANIM_COLOR:-BLUE}"
echo "Preview: ${MANIM_PREVIEW:-false}"

export MANIM_COLOR=${MANIM_COLOR:-BLUE}
export MANIM_QUALITY=${MANIM_QUALITY:-medium}
export MANIM_PREVIEW=${MANIM_PREVIEW:-false}

exec "$@"
