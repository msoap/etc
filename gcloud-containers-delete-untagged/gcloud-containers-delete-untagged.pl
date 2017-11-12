#!/usr/bin/perl

# Delete all untagged images from Google Container Registry.
#
# Usage:
#  gcloud-containers-delete-untagged.pl [--dry-run] us.gcr.io/PROJECT-ID/{repo-name1,repo-name2}
#
# get size of registry:
#   gsutil du -ch 'gs://us.artifacts.PROJECT-ID.appspot.com/' | tail -1
#
# get all image names
#   gsutil ls 'gs://us.artifacts.PROJECT-ID.appspot.com/containers/repositories/library/' | awk -F/ '{print $7}'
#
# used commands from gcloud SDK (https://cloud.google.com/sdk/gcloud/reference/beta/container/images/):
#   gcloud beta container images list-tags us.gcr.io/PROJECT-ID/repo | awk 'NF == 2 && $2 ~ /^[0-9]/ {print $1}'
#   gcloud beta container images delete us.gcr.io/PROJECT-ID/repo@sha256:DIGEST

use warnings;
use strict;

# -----------------------------------------------------------------------------
sub main {
    if (scalar(@ARGV) == 0 || $ARGV[0] eq '--help') {
        print "Usage: $0 [--dry-run] gcr.io/PROJECT-ID/repo1 gcr.io/PROJECT-ID/repo2 ...\n";
        return;
    }

    my $dry_run = 0;
    if ($ARGV[0] eq '--dry-run') {
        $dry_run = 1;
        shift @ARGV;
    }

    for my $repo (@ARGV) {
        print "Try $repo...\n";
        my @digests = grep {length($_) > 1} map {chomp $_; $_} `gcloud beta container images list-tags $repo --limit=100 | awk 'NF == 2 && \$2 ~ /^[0-9]/ {print \$1}'`;
        if (scalar(@digests) == 0) {
            print "Nothing todo\n";
            next;
        }

        my $all_images = join(" ", map {"$repo\@sha256:$_"} @digests);
        my $delete_command = "gcloud beta container images delete $all_images";

        print "$delete_command\n";
        if (! $dry_run) {
            system($delete_command);
        }
    }
}

# -----------------------------------------------------------------------------
main();
