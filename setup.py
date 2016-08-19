import os
from setuptools import setup

setup(
        name = "surfboard_exporter",
        version = "1.0.0",
        author = "Jarod Watkins",
        author_email = "jwatkins@jarodw.com",
        description = ("Arris Surfboard exporter for the Prometheus monitoring system."),
        long_description = ("See https://github.com/ipstatic/surfboard_exporter/blob/master/README.md for documenation."),
        license = "MIT",
        keywords = "prometheus exporter monitoring arris surfboard modem",
        url = "https://github.com/ipstatic/surfboard_exporter",
        scripts = ["scripts/surfboard_exporter"],
        packages = ["surfboard_exporter"],
        test_suite = "tests",
        install_requires = ["prometheus_client>=0.0.14", "requests>=2.6.0", "lxml>=3.2.1"],
        classifiers = [
            "Development Status :: 3 - Alpha",
            "Intended Audience :: Information Technology",
            "Intended Audience :: System Administrators",
            "Topic :: System :: Monitoring",
            "License :: OSI Approved :: MIT License",
        ],
)
