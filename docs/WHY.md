## What is Phoenix? 

Phoenix is a Kubernetes Operator that helps protecting your applications running inside Kubernetes, based on the priciples of Moving Target Defense and actually taking it one step further to implement Automated Moving Target Defense (AMTD).

## What is Automated Moving Target Defense (AMTD)?

In order to understand AMTD let's take a look first on Moving Target Defense (MTD). MTD as one of the game-changing themes, provides a new idea to improve cyberspace security. 

The concept of moving target defense was proposed at first in the U.S. national cyber leap year summit in 2009. In 2012, the report of White House national security council explains the meaning of "moving target", which is systems that move in multiple dimensions to go against attacks and increase system resiliency. In 2014, moving target defense concept is defined as follows by Federal Cybersecurity Research and Development Program.

Moving target defense enables us to create, analyze, evaluate, and deploy mechanisms and strategies that are diverse and that continually shift and change over time to increase complexity and cost for attackers, limit the exposure of vulnerabilities and opportunities for attack, and increase system resiliency.

At its core, MTD incorporates four main elements:

* Proactive cyber defense mechanisms
* Automation to orchestrate movement or change in the attack surface
* The use of deception technologies
* The ability to execute intelligent (preplanned) change decisions

Overall, MTD has the effect of reducing exposed attack surfaces by introducing strategic change, and it increases the cost of reconnaissance and malicious exploitation on the attacker. In other
words, it’s about moving, changing, obfuscating or morphing various aspects of attack surfaces to thwart attacker activities and disrupt the cyber kill chain.

You can find further details [here](https://www.hindawi.com/journals/scn/2018/3759626/).

Although multiple markets and security domains are already using moving target defense (MTD) techniques, automation is the new, and disruptive, frontier. AMTD effectively mitigates many known threats and is likely to mitigate most zero-day exploits within a decade, rotating risks further to humans and business processes.

## Why care about Automated Moving Target Defense in your Kubernetes based operation platforms?

After you carefully tested your application source code with static code analyzators, created minimal container images, set up networking security measures, configured user authentication and authorization it still can happen that an attacker finds a way into your system. 
In the context of Kubernetes AMTD provides a new dimension that you can add to your existing defense vectors. 

Phoenix brings the AMTD principles inside your cluster and watches threats that 3rd party security analytics tools like Falco, Kubearmor, etc. notice and react to these threats by an extendable list of actions. 
You can construct your very own strategy on:
* what applications to watch
* how to react to specific kind of threats

Due to the loosly coupled, extendable architecture of Phoenix you have great flexibility in designing the strategies you need.

* [Getting Started](README.md#getting-started): Get started with Phoenix!
* [Tutorials](README.md#tutorials): Check out the tutorials!

