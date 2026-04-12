import Hero from "./_components/hero";
import Navbar from "./_components/navbar";
import Problem from "./_components/problem";
import Solution from "./_components/solution";
import HowItWorks from "./_components/how-it-works";
import CodeExample from "./_components/code-example";
import Features from "./_components/features";
import WhyItsDifferent from "./_components/why-its-different";
import BetaNotice from "./_components/beta-notice";
import CallToAction from "./_components/cta";
import Footer from "./_components/footer";

export default function HomePage() {
  return (
    <div className="flex flex-col relative w-full">
      <Navbar />
      <main className="flex flex-col">
        <Hero />
        <Problem />
        <Solution />
        <HowItWorks />
        <CodeExample />
        <Features />
        <WhyItsDifferent />
        <BetaNotice />
        <CallToAction />
      </main>
      <Footer />
    </div>
  );
}
