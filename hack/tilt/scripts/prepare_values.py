# SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
# SPDX-FileContributor: enriqueavi@inditex.com
#
# SPDX-License-Identifier: Apache-2.0

#!/usr/bin/env python3
import argparse
import subprocess
import yaml
import os
import shutil
import sys
import time


def wait_for_deployment(namespace, deployment):
    print("[INFO] Waiting 30 seconds for Tilt to update the deployment...")
    time.sleep(30)

    print(f"[INFO] Waiting for deployment '{deployment}' to be available in namespace '{namespace}'...")
    try:
        subprocess.run(
            [
                "kubectl", "wait",
                f"deployment/{deployment}",
                "-n", namespace,
                "--for=condition=available",
                "--timeout=60s"
            ],
            check=True
        )
        print("[INFO] Deployment available.")
    except subprocess.CalledProcessError:
        print("[WARN] The deployment did not become available within the expected time.")


def get_image_from_deployment(namespace, deployment):
    try:
        cmd = [
            "kubectl", "-n", namespace,
            "get", f"deployment/{deployment}",
            "-o", "jsonpath={.spec.template.spec.containers[0].image}"
        ]
        image = subprocess.check_output(cmd, text=True).strip()
        return image if image else None
    except subprocess.CalledProcessError:
        return None

def extract_tag(image):
    if ':' in image:
        # Take everything after the last ':', to support repositories with port
        return image.rsplit(':', 1)[1]
    else:
        return 'dev'

def main():
    parser = argparse.ArgumentParser(description="Prepare values.yaml for Helm with the current image")
    parser.add_argument('--namespace', required=True)
    parser.add_argument('--deployment', required=True)
    parser.add_argument('--values-dev', required=True)
    parser.add_argument('--values', required=True)
    parser.add_argument('--default-image', required=True)
    args = parser.parse_args()

    if not os.path.exists(args.values):
        print(f"[INFO] {args.values} does not exist, copying {args.values_dev} without changes")
        shutil.copyfile(args.values_dev, args.values)
        sys.exit(0)

    wait_for_deployment(args.namespace, args.deployment)
    print(f"[INFO] Getting image from deployment {args.deployment} in namespace {args.namespace}...")

    image = get_image_from_deployment(args.namespace, args.deployment)
    if not image:
        print(f"[WARN] Could not get image from deployment {args.deployment} in namespace {args.namespace}, will use default image")
        image = args.default_image

    tag = extract_tag(image)

    with open(args.values_dev) as f:
        values = yaml.safe_load(f)

    # Ensure path exists
    if 'deployment' not in values:
        values['deployment'] = {}
    if 'image' not in values['deployment']:
        values['deployment']['image'] = {}

    values['deployment']['image']['tag'] = tag

    with open(args.values, 'w') as f:
        yaml.safe_dump(values, f)

    print(f"[INFO] Updated {args.values} with image tag: {tag}")

if __name__ == '__main__':
    main()
