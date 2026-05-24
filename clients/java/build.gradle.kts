plugins {
  `java-library`
  `maven-publish`
  checkstyle
  id("com.diffplug.spotless") version "8.0.0"
}

group = "com.retab"
version = "0.0.1"

java {
  sourceCompatibility = JavaVersion.VERSION_11
  targetCompatibility = JavaVersion.VERSION_11
  withSourcesJar()
  withJavadocJar()
}

dependencies {
  api("com.fasterxml.jackson.core:jackson-databind:2.17.2")
  api("com.fasterxml.jackson.datatype:jackson-datatype-jsr310:2.17.2")
  testImplementation("org.junit.jupiter:junit-jupiter:5.11.4")
  testRuntimeOnly("org.junit.platform:junit-platform-launcher:1.11.4")
}

spotless {
  java {
    googleJavaFormat("1.35.0")
  }
}

checkstyle {
  configFile = file("checkstyle.xml")
  toolVersion = "10.21.1"
  isIgnoreFailures = false
}

tasks.withType<Test>().configureEach {
  useJUnitPlatform()
}

tasks.withType<Javadoc>().configureEach {
  (options as StandardJavadocDocletOptions).addStringOption("Xdoclint:none", "-quiet")
}

tasks.check {
  dependsOn(tasks.spotlessCheck)
}

publishing {
  publications {
    create<MavenPublication>("mavenJava") {
      artifactId = "retab"
      from(components["java"])
      pom {
        name.set("Retab")
        description.set("Generated Retab Java SDK.")
        url.set("https://retab.com")
        licenses {
          license {
            name.set("MIT License")
            url.set("https://opensource.org/license/mit")
          }
        }
        developers {
          developer {
            organization.set("Retab")
            organizationUrl.set("https://retab.com")
          }
        }
        scm {
          url.set("https://github.com/retab-dev/retab")
          connection.set("scm:git:https://github.com/retab-dev/retab.git")
          developerConnection.set("scm:git:ssh://git@github.com/retab-dev/retab.git")
        }
      }
    }
  }
  repositories {
    maven {
      name = "localBuild"
      url = uri(layout.buildDirectory.dir("repository"))
    }
  }
}
